package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"spark/internal/database"
	"spark/internal/models"
)

// UserContextKey is the key for the user value in the request context.
const UserContextKey = contextKey("user")

// Protect is a middleware to verify a user's JWT token.
func Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default_secret_change_me"
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSONError(w, "Not authorized, no token", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			writeJSONError(w, "Not authorized, token format is 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			writeJSONError(w, "Not authorized, token failed", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["id"] == nil {
			writeJSONError(w, "Not authorized, invalid token claims", http.StatusUnauthorized)
			return
		}

		id := claims["id"]
		var userType string
		// Try to get role/userType from claims
		if role, ok := claims["role"].(string); ok {
			userType = role
		}

		// SPECIAL CASE: Handle the admin user
		if id == "admin" && userType == "Admin" {
			// Create a mock admin user.
			adminUser := &models.User{
				UserType: "Admin",
			}
			ctx := context.WithValue(r.Context(), UserContextKey, adminUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// STANDARD CASE: Handle regular users with numeric IDs
		var userID uint
		switch v := id.(type) {
		case float64:
			userID = uint(v)
		default:
			writeJSONError(w, "Unauthorized: Invalid user ID in token.", http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := database.DB.Omit("password").First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeJSONError(w, "Not authorized, user not found", http.StatusUnauthorized)
				return
			}
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserFromContext retrieves the user from the context.
func getUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// IsAdmin checks if the user is an Admin.
func IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if user == nil || user.UserType != "Admin" {
			writeJSONError(w, "Access forbidden. You must be an admin.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IsDriver checks if the user is a Driver.
func IsDriver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if user == nil || user.UserType != "Driver" {
			writeJSONError(w, "Access forbidden. You must be a Spark Partner.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IsCustomer checks if the user is a Customer.
func IsCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if user == nil || user.UserType != "Customer" {
			writeJSONError(w, "Access forbidden. You must be a customer.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
