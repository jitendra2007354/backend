package services

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	userSockets  = sync.Map{} // map[uint]*websocket.Conn
	adminSockets = sync.Map{}
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func InitSocketService() {
	fmt.Println("Socket service initialized")
	// In a real app, you would register a route like /ws here or in routes
	http.HandleFunc("/ws", handleWebSocket)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Basic handshake logic (simplified)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// Assume user ID is passed via query param for this simple example
	// userID := r.URL.Query().Get("userId")
	// userSockets.Store(userID, conn)
	defer conn.Close()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func SendMessageToUser(userID uint, event string, data interface{}) {
	if conn, ok := userSockets.Load(userID); ok {
		ws := conn.(*websocket.Conn)
		ws.WriteJSON(map[string]interface{}{"event": event, "data": data})
	} else {
		fmt.Printf("User %d not connected. Event: %s\n", userID, event)
	}
}

func SendMessageToRoom(room, event string, data interface{}) {
	// In a real implementation, you would maintain a map[room_name][]*websocket.Conn
	// For this simple version, we iterate all users (inefficient but functional for small scale)
	userSockets.Range(func(key, value interface{}) bool {
		ws := value.(*websocket.Conn)
		ws.WriteJSON(map[string]interface{}{"event": event, "data": data})
		return true
	})
}

func SendMessageToAdminRoom(event string, data interface{}) {
	fmt.Printf("Sending %s to admins: %v\n", event, data)
}

func BroadcastSponsorNotification(target, title, message string, notifID uint) (struct{DriverCount, CustomerCount int}, error) {
	// Mock implementation
	return struct{DriverCount, CustomerCount int}{10, 20}, nil
}
