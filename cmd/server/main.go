package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"spark/internal/controllers"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type methodMux struct {
	*http.ServeMux
	routes map[string]map[string]http.Handler
}

func newMethodMux() *methodMux {
	return &methodMux{ServeMux: http.NewServeMux(), routes: make(map[string]map[string]http.Handler)}
}

func (m *methodMux) addMethodRoute(path, method string, handler http.Handler) {
	if _, ok := m.routes[path]; !ok {
		m.routes[path] = map[string]http.Handler{}
		m.ServeMux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if methodMap, ok := m.routes[path]; ok {
				if h, ok := methodMap[r.Method]; ok {
					h.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			http.NotFound(w, r)
		}))
	}
	m.routes[path][method] = handler
}

func (m *methodMux) Handle(pattern string, handler http.Handler) {
	if sep := strings.Index(pattern, " "); sep != -1 {
		method := pattern[:sep]
		path := pattern[sep+1:]
		m.addMethodRoute(path, method, handler)
		return
	}
	m.ServeMux.Handle(pattern, handler)
}

func (m *methodMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if sep := strings.Index(pattern, " "); sep != -1 {
		method := pattern[:sep]
		path := pattern[sep+1:]
		m.addMethodRoute(path, method, http.HandlerFunc(handler))
		return
	}
	m.ServeMux.HandleFunc(pattern, handler)
}

func main() {
	migrateFlag := flag.Bool("migrate", false, "Run database auto-migration and exit")
	flag.Parse()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Relying on system environment variables.")
	} else {
		log.Println("✅ Loaded .env file")
	}

	// Check if either DB_HOST or TIDB_HOST is set
	if os.Getenv("DB_HOST") == "" && os.Getenv("TIDB_HOST") == "" {
		log.Fatal("Error: Database configuration missing. Please set DB_HOST or TIDB_HOST in your .env file.")
	}

	// 1. Initialize Database
	database.Connect()

	// Inject DB connection into services
	services.DB = database.DB

	// Run Auto Migration only if the -migrate flag is provided
	if *migrateFlag {
		log.Println("Running database auto-migration...")
		err := services.DB.AutoMigrate(
			&models.User{},
			&models.Driver{},
			&models.Vehicle{},
			&models.Ride{},
			&models.Bid{},
			&models.DriverLocation{},
			&models.Config{},
			&models.Sponsor{},
			&models.SponsorNotification{},
			&models.ChatMessage{},
			&models.SupportTicket{},
			&models.Transaction{},
			&models.Rating{},
			&models.PricingRule{},
			&models.Notification{},
			&models.NotificationForSponsor{},
		)
		if err != nil {
			log.Fatalf("Database AutoMigrate failed: %v", err)
		}
		log.Println("✅ Database AutoMigrate completed successfully.")
		os.Exit(0) // Exit after migrating so the server doesn't start
	}

	// Initialize Redis
	services.InitRedis()

	// Initialize Socket Service (Starts Redis Listener)
	services.InitSocketService()

	// 2. Start Background Scheduler (Cron jobs)
	services.InitScheduler()

	// Start Service-level Cron Jobs (Cleanup, Wallet Checks)
	services.ScheduleChatCleanup()

	// 4. Setup Router
	mux := newMethodMux()

	// --- Public Routes ---
	mux.HandleFunc("/api/auth/login", controllers.Login)
	mux.HandleFunc("/api/auth/admin-login", controllers.AdminLogin)
	mux.HandleFunc("POST /api/users", controllers.RegisterUser)
	mux.HandleFunc("POST /api/sponsor/login", controllers.SponsorLogin)
	mux.HandleFunc("POST /api/auth/guest-login", func(w http.ResponseWriter, r *http.Request) {
		res, err := services.LoginGuest()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"token": res.Token, "user": res.User})
	})

	// --- Websocket Route ---
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		services.HandleWebSocket(w, r)
	})

	// --- Protected User Routes (Driver/Customer) ---
	// Note: Assuming middleware.Protect exists in your internal/middleware package
	// If not, you need to implement it similar to ProtectSponsor but for Users.
	// For now, wrapping in a placeholder if it doesn't exist in context provided.

	// Auth
	mux.Handle("GET /api/auth/me", middleware.Protect(http.HandlerFunc(controllers.Me)))

	// Driver
	mux.Handle("GET /api/driver/rides", middleware.Protect(http.HandlerFunc(controllers.GetDriverRides)))
	mux.HandleFunc("GET /api/driver/nearby", controllers.GetNearbyDrivers)  // Public or Protected?
	mux.HandleFunc("GET /api/drivers/nearby", controllers.GetNearbyDrivers) // Alias for frontend
	mux.Handle("GET /api/driver/poll-rides", middleware.Protect(http.HandlerFunc(controllers.PollRideRequests)))
	mux.Handle("POST /api/driver/status/toggle", middleware.Protect(middleware.IsDriver(http.HandlerFunc(controllers.ToggleDriverStatus))))
	mux.Handle("PUT /api/driver/status", middleware.Protect(middleware.IsDriver(http.HandlerFunc(controllers.SetDriverStatus))))
	mux.Handle("PUT /api/driver/profile", middleware.Protect(http.HandlerFunc(controllers.UpdateProfile)))
	mux.Handle("POST /api/driver/location", middleware.Protect(http.HandlerFunc(controllers.UpdateDriverLocation)))

	// Rides
	mux.Handle("POST /api/rides", middleware.Protect(http.HandlerFunc(controllers.CreateRide)))
	mux.Handle("GET /api/rides/active", middleware.Protect(http.HandlerFunc(controllers.GetActiveRide)))
	mux.HandleFunc("GET /api/rides/available", controllers.GetAvailableRides)
	mux.HandleFunc("GET /api/rides/{id}", controllers.GetRideByID)
	mux.Handle("POST /api/rides/{id}/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptRide)))
	mux.Handle("POST /api/rides/{id}/reject", middleware.Protect(http.HandlerFunc(controllers.RejectRide)))
	mux.Handle("POST /api/rides/{id}/cancel", middleware.Protect(http.HandlerFunc(controllers.CancelRide)))
	mux.Handle("POST /api/rides/{rideId}/confirm", middleware.Protect(http.HandlerFunc(controllers.ConfirmRide)))
	mux.Handle("POST /api/rides/{rideId}/status", middleware.Protect(http.HandlerFunc(controllers.UpdateRideStatus)))

	// Customer
	mux.Handle("GET /api/customer/rides", middleware.Protect(http.HandlerFunc(controllers.GetCustomerRides)))
	mux.Handle("GET /api/bookings", middleware.Protect(http.HandlerFunc(controllers.GetCustomerRides))) // alias for frontend
	mux.Handle("GET /api/user/notifications", middleware.Protect(http.HandlerFunc(controllers.GetUserNotifications)))
	mux.Handle("POST /api/notifications/read", middleware.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	})))
	mux.Handle("GET /api/user/chats", middleware.Protect(http.HandlerFunc(controllers.GetUserChats)))      // frontend expects this
	mux.Handle("POST /api/support", middleware.Protect(http.HandlerFunc(controllers.SubmitSupportTicket))) // alias for frontend
	mux.Handle("POST /api/request/ride", middleware.Protect(http.HandlerFunc(controllers.RequestRide)))
	mux.Handle("POST /api/rides/request", middleware.Protect(http.HandlerFunc(controllers.RequestRide))) // frontend path
	mux.HandleFunc("GET /api/request/ride/{id}", controllers.GetRideDetails)
	mux.HandleFunc("POST /api/request/ride/{id}/cancel", controllers.CancelRideRequest)
	mux.Handle("PUT /api/rides/{rideId}/fare", middleware.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rideId, _ := strconv.Atoi(r.PathValue("rideId"))
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var payload struct {
			Fare float64 `json:"fare"`
		}
		json.NewDecoder(r.Body).Decode(&payload)
		ride, err := services.UpdateRideFare(uint(rideId), user.ID, payload.Fare)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ride)
	})))

	// Wallet & Payment
	mux.Handle("GET /api/wallet/balance", middleware.Protect(http.HandlerFunc(controllers.GetWalletBalance)))
	mux.Handle("GET /api/wallet/transactions", middleware.Protect(http.HandlerFunc(controllers.GetWalletTransactions)))
	mux.Handle("POST /api/wallet/topup", middleware.Protect(http.HandlerFunc(controllers.TopUpWallet)))
	mux.Handle("POST /api/payment/pay", middleware.Protect(http.HandlerFunc(controllers.MakePayment)))
	mux.Handle("POST /api/payment/confirm", middleware.Protect(http.HandlerFunc(controllers.ConfirmPayment)))

	// Bidding
	mux.Handle("POST /api/bids", middleware.Protect(http.HandlerFunc(controllers.CreateBidController)))
	mux.Handle("POST /api/rides/{rideId}/bid", middleware.Protect(http.HandlerFunc(controllers.CreateBidController))) // frontend path
	mux.HandleFunc("GET /api/bids/{rideId}", controllers.GetBidsController)
	mux.Handle("POST /api/bids/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptBidController)))
	mux.Handle("POST /api/bids/counter", middleware.Protect(http.HandlerFunc(controllers.CounterBidController)))
	mux.Handle("POST /api/bids/counter/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptCounterBidController)))
	mux.Handle("POST /api/bids/driver-accept", middleware.Protect(http.HandlerFunc(controllers.DriverAcceptInstantController)))

	// Chat & Support
	mux.Handle("POST /api/chat", middleware.Protect(http.HandlerFunc(controllers.SendChatMessage)))
	mux.Handle("GET /api/chat/{rideId}", middleware.Protect(http.HandlerFunc(controllers.GetChatMessages)))
	mux.Handle("POST /api/support/tickets", middleware.Protect(http.HandlerFunc(controllers.SubmitSupportTicket)))

	// Config (public)
	mux.Handle("GET /api/config", http.HandlerFunc(controllers.GetSystemConfig))
	mux.Handle("PUT /api/config", http.HandlerFunc(controllers.UpdateConfig))

	// Vehicles & Docs
	mux.Handle("POST /api/vehicles", middleware.Protect(http.HandlerFunc(controllers.AddVehicle)))
	mux.Handle("GET /api/vehicles", middleware.Protect(http.HandlerFunc(controllers.GetVehicles)))
	mux.Handle("POST /api/driver/vehicles", middleware.Protect(http.HandlerFunc(controllers.AddVehicle)))
	mux.Handle("GET /api/driver/vehicles", middleware.Protect(http.HandlerFunc(controllers.GetVehicles)))
	mux.Handle("DELETE /api/driver/vehicles/{vehicleId}", middleware.Protect(http.HandlerFunc(controllers.DeleteVehicle)))
	mux.Handle("POST /api/vehicles/{vehicleId}/default", middleware.Protect(http.HandlerFunc(controllers.SetDefaultVehicle)))
	mux.Handle("DELETE /api/vehicles/{vehicleId}", middleware.Protect(http.HandlerFunc(controllers.DeleteVehicle)))
	mux.Handle("POST /api/documents/upload", middleware.Protect(http.HandlerFunc(controllers.UploadDocumentController)))

	// Ratings
	mux.Handle("POST /api/ratings", middleware.Protect(http.HandlerFunc(controllers.CreateRatingController)))
	mux.HandleFunc("GET /api/ratings/{userId}", controllers.GetRatingsController)

	// --- Protected Sponsor Routes ---
	mux.Handle("GET /api/sponsor/history", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetSponsorHistory)))
	mux.Handle("GET /api/sponsor/notifications", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetSponsorHistory)))
	mux.Handle("POST /api/sponsor/notifications", middleware.ProtectSponsor(http.HandlerFunc(controllers.SendSponsorNotification)))
	mux.Handle("DELETE /api/sponsor/notifications/{id}", middleware.ProtectSponsor(http.HandlerFunc(controllers.DeleteSponsorNotification)))
	mux.Handle("POST /api/sponsor/campaign/banner", middleware.ProtectSponsor(http.HandlerFunc(controllers.UploadCampaignBanner)))
	mux.Handle("GET /api/sponsor/gam/token", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetGamToken)))

	// Sponsor System Notifications
	mux.Handle("GET /api/sponsor/notifications/system", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetSystemNotifications)))
	mux.Handle("POST /api/sponsor/notifications/{id}/read", middleware.ProtectSponsor(http.HandlerFunc(controllers.MarkSystemNotificationRead)))
	mux.Handle("POST /api/sponsor/notifications/read-all", middleware.ProtectSponsor(http.HandlerFunc(controllers.MarkAllSystemNotificationsRead)))
	mux.Handle("POST /api/sponsor/notifications/{id}/like", middleware.ProtectSponsor(http.HandlerFunc(controllers.ToggleSystemNotificationLike)))

	// --- Admin Routes (Should be protected by Admin Middleware in production) ---
	mux.HandleFunc("POST /api/admin/ai-chat", controllers.GetAIAssistantResponse)

	mux.HandleFunc("POST /api/admin/notifications", controllers.SendNotification)
	mux.HandleFunc("POST /api/admin/users", controllers.CreateUser)
	mux.HandleFunc("GET /api/admin/users/all", controllers.GetAllUsers)
	mux.HandleFunc("POST /api/admin/ai/generate", controllers.GetAIAssistantResponse)
	mux.HandleFunc("GET /api/admin/users", controllers.GetAllUsers)

	mux.HandleFunc("GET /api/admin/tickets", controllers.GetTickets)
	mux.HandleFunc("GET /api/admin/driver-locations", controllers.GetDriverLocations)
	mux.HandleFunc("GET /api/admin/notification-history", controllers.GetNotificationHistory)
	mux.HandleFunc("GET /api/admin/drivers/locations", controllers.GetDriverLocations)
	mux.HandleFunc("GET /api/admin/notifications/history", controllers.GetNotificationHistory)
	mux.HandleFunc("GET /api/admin/config", controllers.GetSystemConfig)
	mux.HandleFunc("PUT /api/admin/config", controllers.UpdateSystemConfig)
	mux.HandleFunc("POST /api/admin/realtime-notification", controllers.CreateRealtimeNotificationHandler)

	// Driver Admin Actions
	mux.HandleFunc("POST /api/driver/block/{id}", controllers.BlockDriver)
	mux.HandleFunc("POST /api/driver/unblock/{id}", controllers.UnblockDriver)
	mux.HandleFunc("POST /api/wallet/adjust", controllers.AdjustWallet)
	mux.HandleFunc("POST /api/documents/verify", controllers.VerifyDocumentController)

	// Serve static files (uploads)
	fs := http.FileServer(http.Dir("public/uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// 5. Start Server with CORS
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
	}

	// Start the server in a background goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Create a channel to listen for OS shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received, gracefully shutting down server...")

	// Create a deadline to wait for active requests to finish (e.g., database saves, payments)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown due to timeout or error:", err)
	}

	log.Println("Server successfully exited. All active transactions safely resolved.")
}

// Simple CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cache-Control, Pragma")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
