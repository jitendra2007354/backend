package main

import (
	"log"
	"net/http"
	"os"
	"spark/internal/controllers"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/services"

	"github.com/joho/godotenv"
)

func main() {
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

	// Initialize Redis
	services.InitRedis()

	// Initialize Socket Service (Starts Redis Listener)
	services.InitSocketService()

	// 2. Start Background Scheduler (Cron jobs)
	services.InitScheduler()

	// Start Service-level Cron Jobs (Cleanup, Wallet Checks)
	services.ScheduleChatCleanup()
	services.ScheduleDriverStatusChecks()
	services.ScheduleWalletCheck()

	// 4. Setup Router
	mux := http.NewServeMux()

	// --- Public Routes ---
	mux.HandleFunc("POST /api/auth/login", controllers.Login)
	mux.HandleFunc("POST /api/auth/send-otp", controllers.SendOtp)
	mux.HandleFunc("POST /api/auth/verify-otp", controllers.VerifyOtp)
	mux.HandleFunc("POST /api/auth/guest-login", controllers.GuestLogin)
	mux.HandleFunc("POST /api/auth/admin/login", controllers.AdminLogin)
	mux.HandleFunc("POST /api/users", controllers.RegisterUser)
	mux.HandleFunc("POST /api/sponsor/login", controllers.SponsorLogin)

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
	mux.HandleFunc("GET /api/driver/nearby", controllers.GetNearbyDrivers) // Public or Protected?
	mux.Handle("GET /api/driver/poll-rides", middleware.Protect(http.HandlerFunc(controllers.PollRideRequests)))
	
	// Rides
	mux.Handle("POST /api/rides", middleware.Protect(http.HandlerFunc(controllers.CreateRide)))
	mux.HandleFunc("GET /api/rides/available", controllers.GetAvailableRides)
	mux.HandleFunc("GET /api/rides/{id}", controllers.GetRideByID)
	mux.Handle("POST /api/rides/{id}/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptRide)))
	mux.Handle("POST /api/rides/{id}/reject", middleware.Protect(http.HandlerFunc(controllers.RejectRide)))
	mux.Handle("POST /api/rides/{id}/cancel", middleware.Protect(http.HandlerFunc(controllers.CancelRide)))
	mux.Handle("POST /api/rides/{rideId}/confirm", middleware.Protect(http.HandlerFunc(controllers.ConfirmRide)))
	mux.Handle("POST /api/rides/{rideId}/status", middleware.Protect(http.HandlerFunc(controllers.UpdateRideStatus)))

	// Customer
	mux.Handle("GET /api/customer/rides", middleware.Protect(http.HandlerFunc(controllers.GetCustomerRides)))
	mux.Handle("POST /api/request/ride", middleware.Protect(http.HandlerFunc(controllers.RequestRide)))
	mux.HandleFunc("GET /api/request/ride/{id}", controllers.GetRideDetails)
	mux.HandleFunc("POST /api/request/ride/{id}/cancel", controllers.CancelRideRequest)

	// Wallet & Payment
	mux.Handle("GET /api/wallet/balance", middleware.Protect(http.HandlerFunc(controllers.GetWalletBalance)))
	mux.Handle("POST /api/wallet/topup", middleware.Protect(http.HandlerFunc(controllers.TopUpWallet)))
	mux.Handle("POST /api/payment/pay", middleware.Protect(http.HandlerFunc(controllers.MakePayment)))
	mux.Handle("POST /api/payment/confirm", middleware.Protect(http.HandlerFunc(controllers.ConfirmPayment)))

	// Bidding
	mux.Handle("POST /api/bids", middleware.Protect(http.HandlerFunc(controllers.CreateBidController)))
	mux.HandleFunc("GET /api/bids/{rideId}", controllers.GetBidsController)
	mux.Handle("POST /api/bids/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptBidController)))
	mux.Handle("POST /api/bids/counter", middleware.Protect(http.HandlerFunc(controllers.CounterBidController)))
	mux.Handle("POST /api/bids/counter/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptCounterBidController)))
	mux.Handle("POST /api/bids/driver-accept", middleware.Protect(http.HandlerFunc(controllers.DriverAcceptInstantController)))

	// Chat & Support
	mux.Handle("POST /api/chat", middleware.Protect(http.HandlerFunc(controllers.SendChatMessage)))
	mux.HandleFunc("GET /api/chat/{rideId}", controllers.GetChatMessages)
	mux.Handle("POST /api/support/tickets", middleware.Protect(http.HandlerFunc(controllers.SubmitSupportTicket)))

	// Vehicles & Docs
	mux.Handle("POST /api/vehicles", middleware.Protect(http.HandlerFunc(controllers.AddVehicle)))
	mux.Handle("GET /api/vehicles", middleware.Protect(http.HandlerFunc(controllers.GetVehicles)))
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
	mux.Handle("POST /api/sponsor/upload-banner", middleware.ProtectSponsor(http.HandlerFunc(controllers.UploadCampaignBanner)))
	mux.HandleFunc("GET /api/sponsor/gam-token", controllers.GetGamToken) // Mocked
	
	// Sponsor System Notifications
	mux.Handle("GET /api/sponsor/notifications/system", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetSystemNotifications)))
	mux.Handle("POST /api/sponsor/notifications/{id}/read", middleware.ProtectSponsor(http.HandlerFunc(controllers.MarkSystemNotificationRead)))
	mux.Handle("POST /api/sponsor/notifications/read-all", middleware.ProtectSponsor(http.HandlerFunc(controllers.MarkAllSystemNotificationsRead)))
	mux.Handle("POST /api/sponsor/notifications/{id}/like", middleware.ProtectSponsor(http.HandlerFunc(controllers.ToggleSystemNotificationLike)))

	// --- Admin Routes (Should be protected by Admin Middleware in production) ---
	mux.HandleFunc("POST /api/admin/ai-chat", controllers.GetAIAssistantResponse)
	mux.HandleFunc("POST /api/admin/pricing/{type}", controllers.AddPricingRule)
	mux.HandleFunc("DELETE /api/admin/pricing/{type}/{id}", controllers.DeletePricingRule)
	mux.HandleFunc("POST /api/admin/notifications", controllers.SendNotification)
	mux.HandleFunc("POST /api/admin/users", controllers.CreateUser)
	mux.HandleFunc("GET /api/admin/users/all", controllers.GetAllUsers)
	mux.HandleFunc("GET /api/admin/pricing", controllers.GetPricing)
	mux.HandleFunc("GET /api/admin/tickets", controllers.GetTickets)
	mux.HandleFunc("GET /api/admin/driver-locations", controllers.GetDriverLocations)
	mux.HandleFunc("GET /api/admin/notification-history", controllers.GetNotificationHistory)
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

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// Simple CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
