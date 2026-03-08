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
	"spark/internal/models" // NOTE: Replace 'spark' with your actual Go module name
)

// A private key for context that is unexported to prevent collisions.
type contextKey string

// SponsorContextKey is the key for the sponsor value in the request context.
const SponsorContextKey = contextKey("sponsor")

// writeJSONError is a helper to write a JSON error message.
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"message":"%s"}`, message)
}

// ProtectSponsor is a middleware to verify a sponsor's JWT token and add the sponsor to the request context.
func ProtectSponsor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtSecret := os.Getenv("JWT_SPONSOR_SECRET")
		if jwtSecret == "" {
			jwtSecret = "a_very_secure_secret_for_sponsors"
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

		sponsorID := uint(claims["id"].(float64))

		var sponsor models.Sponsor
		if err := database.DB.Omit("password").First(&sponsor, sponsorID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeJSONError(w, "Not authorized, sponsor not found", http.StatusUnauthorized)
				return
			}
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), SponsorContextKey, &sponsor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}