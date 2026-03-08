package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"spark/internal/models"
)

// UserData holds information about a connected websocket user.
type UserData struct {
	ID            uint
	UserType      string
	DriverID      uint
	LastDBUpdate  int64
	LastDBLat     float64
	LastDBLng     float64
}

var (
	userSockets      = sync.Map{} // map[uint]*websocket.Conn
	adminSockets     = sync.Map{} // map[*websocket.Conn]bool
	roomSubscriptions = struct {
		sync.RWMutex
		m map[string]map[*websocket.Conn]bool
	}{m: make(map[string]map[*websocket.Conn]bool)}
	connData         = sync.Map{} // map[*websocket.Conn]*UserData
	upgrader         = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// broadcastMsg is the payload sent over Redis Pub/Sub
type broadcastMsg struct {
	Type   string      `json:"type"`   // "user", "admin", "room"
	Target string      `json:"target"` // userID, or roomName
	Event  string      `json:"event"`
	Data   interface{} `json:"data"`
}

// InitSocketService registers the /ws handler.
func InitSocketService() {
	// Start Redis subscriber if available
	if RedisClient != nil {
		go func() {
			pubsub := RedisClient.Subscribe(redisCtx, "spark:broadcast")
			defer pubsub.Close()
			ch := pubsub.Channel()
			for msg := range ch {
				var bm broadcastMsg
				if err := json.Unmarshal([]byte(msg.Payload), &bm); err == nil {
					switch bm.Type {
					case "user":
						var uid uint
						fmt.Sscanf(bm.Target, "%d", &uid)
						sendToUserLocal(uid, bm.Event, bm.Data)
					case "admin":
						sendToAdminRoomLocal(bm.Event, bm.Data)
					case "room":
						publishToRoomLocal(bm.Target, bm.Event, bm.Data)
					}
				}
			}
		}()
	}
	fmt.Println("Socket service initialized")
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// authenticate during upgrade using JWT token query parameter
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("{\"error\":\"missing token\"}"))
		conn.Close()
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret_change_me"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		conn.WriteMessage(websocket.TextMessage, []byte("{\"error\":\"invalid token\"}"))
		conn.Close()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		conn.Close()
		return
	}

	var ud UserData
	ud.ID = uint(claims["id"].(float64))
	ud.UserType, _ = claims["userType"].(string)

	// resolve driver ID once if needed
	if ud.UserType == "Driver" {
		var driver models.Driver
		if err := DB.Where("user_id = ?", ud.ID).First(&driver).Error; err == nil {
			ud.DriverID = driver.ID
		}
	}

	// register socket
	registerConnection(conn, &ud)

	defer func() {
		removeConnection(conn)
		conn.Close()
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg struct {
			Event string                 `json:"event"`
			Data  map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		handleSocketEvent(conn, &ud, msg.Event, msg.Data)
	}
}

func registerConnection(conn *websocket.Conn, ud *UserData) {
	connData.Store(conn, ud)
	if ud.UserType == "Admin" {
		adminSockets.Store(conn, true)
	} else {
		userSockets.Store(ud.ID, conn)
	}
}

func removeConnection(conn *websocket.Conn) {
	if v, ok := connData.Load(conn); ok {
		ud := v.(*UserData)
		if ud.UserType == "Admin" {
			adminSockets.Delete(conn)
		} else {
			userSockets.Delete(ud.ID)
			if ud.DriverID != 0 {
				RemoveDriverLocation(ud.DriverID)
			}
		}
		// remove from all rooms
		roomSubscriptions.Lock()
		for _, conns := range roomSubscriptions.m {
			if conns[conn] {
				delete(conns, conn)
			}
		}
		roomSubscriptions.Unlock()
	}
	connData.Delete(conn)
}

func subscribeRoom(conn *websocket.Conn, room string) {
	roomSubscriptions.Lock()
	defer roomSubscriptions.Unlock()
	if _, ok := roomSubscriptions.m[room]; !ok {
		roomSubscriptions.m[room] = make(map[*websocket.Conn]bool)
	}
	roomSubscriptions.m[room][conn] = true
}

// publishToRoom broadcasts to all instances via Redis, or locally if Redis is missing
func publishToRoom(room, event string, data interface{}) {
	if RedisClient != nil {
		payload, _ := json.Marshal(broadcastMsg{Type: "room", Target: room, Event: event, Data: data})
		RedisClient.Publish(redisCtx, "spark:broadcast", payload)
	} else {
		publishToRoomLocal(room, event, data)
	}
}

// publishToRoomLocal sends to connections on THIS server instance
func publishToRoomLocal(room, event string, data interface{}) {
	roomSubscriptions.RLock()
	defer roomSubscriptions.RUnlock()
	if conns, ok := roomSubscriptions.m[room]; ok {
		for c := range conns {
			c.WriteJSON(map[string]interface{}{"event": event, "data": data})
		}
	}
}

func handleSocketEvent(conn *websocket.Conn, ud *UserData, event string, data map[string]interface{}) {
	switch event {
	case "joinRideBidding":
		if rideID, ok := data["rideId"]; ok {
			subscribeRoom(conn, fmt.Sprintf("ride-%v", rideID))
		}

	case "request_ride":
		// convert payload
		pickup := toGeoPoint(data["pickupLocation"])
		dropoff := toGeoPoint(data["dropoffLocation"])
		distance := data["distance"].(float64)
		duration := data["duration"].(float64)
		ride, err := CreateRideService(ud.ID, pickup, dropoff, data["vehicleType"].(string), data["fare"].(float64), distance, duration)
		if err == nil {
			conn.WriteJSON(map[string]interface{}{"event": "request_ride_success", "data": map[string]interface{}{"rideId": ride.ID}})
		}

	case "placeBid":
		if ud.UserType == "Driver" {
			rideID := uint(data["rideId"].(float64))
			amount := data["amount"].(float64)
			bid, err := CreateBid(rideID, ud.ID, amount, uint(data["customerId"].(float64)))
			if err == nil {
				publishToRoom(fmt.Sprintf("ride-%d", bid.RideID), "bidUpdate", bid)
			}
		}

	case "driver_accept_offer":
		if ud.UserType == "Driver" {
			rideID := uint(data["rideId"].(float64))
			HandleDriverAcceptChain(rideID, ud.ID)
		}

	case "driver_reject_offer":
		if ud.UserType == "Driver" {
			rideID := uint(data["rideId"].(float64))
			HandleDriverReject(rideID, ud.ID)
		}

	case "complete_ride":
		if ud.UserType == "Driver" {
			rideID := uint(data["rideId"].(float64))
			HandleRideCompletion(rideID)
		}

	case "cancel_ride":
		if ud.UserType == "Customer" || ud.UserType == "Driver" {
			rideID := uint(data["rideId"].(float64))
			HandleRideCancellation(rideID, ud.UserType)
		}

	case "joinChat":
		subscribeRoom(conn, fmt.Sprintf("chat-%v", data["rideId"]))

	case "sendMessage":
		rideID := uint(data["rideId"].(float64))
		message := data["message"].(string)
		msg, _ := SaveChatMessage(rideID, ud.ID, message)
		publishToRoom(fmt.Sprintf("chat-%v", rideID), "newMessage", msg)

	case "updateLocation":
		if ud.UserType == "Driver" {
			lat := data["lat"].(float64)
			lng := data["lng"].(float64)
			now := time.Now().Unix()
			shouldPersist := (now-ud.LastDBUpdate > 10) || (abs(lat-ud.LastDBLat) > 0.0004 || abs(lng-ud.LastDBLng) > 0.0004)
			if shouldPersist {
				ud.LastDBUpdate = now
				ud.LastDBLat = lat
				ud.LastDBLng = lng
			}
			if ud.DriverID != 0 {
				UpdateDriverLocation(ud.DriverID, lat, lng, shouldPersist)
			}
		}

	case "subscribe_admin_live_map":
		if ud.UserType == "Admin" {
			locs, _ := GetAllOnlineDriverLocations()
			conn.WriteJSON(map[string]interface{}{"event": "admin_live_map_data", "data": map[string]interface{}{"locations": locs}})
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func toGeoPoint(raw interface{}) models.GeoPoint {
	m, _ := raw.(map[string]interface{})
	return models.GeoPoint{Type: "Point", Coordinates: []float64{m["longitude"].(float64), m["latitude"].(float64)}}
}

// Public helpers (used by other packages)
func RegisterUserSocket(userID uint, conn *websocket.Conn) {
	userSockets.Store(userID, conn)
}

func RemoveUserSocket(userID uint) {
	userSockets.Delete(userID)
}

func RegisterAdminSocket(conn *websocket.Conn) {
	adminSockets.Store(conn, true)
}

func RemoveAdminSocket(conn *websocket.Conn) {
	adminSockets.Delete(conn)
}

// Re-export for legacy naming
func PublishToRoom(room, event string, data interface{}) {
	publishToRoom(room, event, data)
}

func SendMessageToUser(userID uint, event string, data interface{}) {
	if RedisClient != nil {
		payload, _ := json.Marshal(broadcastMsg{Type: "user", Target: fmt.Sprintf("%d", userID), Event: event, Data: data})
		RedisClient.Publish(redisCtx, "spark:broadcast", payload)
	} else {
		sendToUserLocal(userID, event, data)
	}
}

func sendToUserLocal(userID uint, event string, data interface{}) {
	if v, ok := userSockets.Load(userID); ok {
		conn := v.(*websocket.Conn)
		conn.WriteJSON(map[string]interface{}{"event": event, "data": data})
	}
}

func SendMessageToAdminRoom(event string, data interface{}) {
	if RedisClient != nil {
		payload, _ := json.Marshal(broadcastMsg{Type: "admin", Event: event, Data: data})
		RedisClient.Publish(redisCtx, "spark:broadcast", payload)
	} else {
		sendToAdminRoomLocal(event, data)
	}
}

func sendToAdminRoomLocal(event string, data interface{}) {
	adminSockets.Range(func(key, value interface{}) bool {
		conn := key.(*websocket.Conn)
		conn.WriteJSON(map[string]interface{}{"event": event, "data": data})
		return true
	})
}