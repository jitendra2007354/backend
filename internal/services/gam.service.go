package services

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

func GetGoogleAdManagerToken() (string, error) {
	ctx := context.Background()
	// Requires GOOGLE_APPLICATION_CREDENTIALS env var pointing to json key file
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/dfp")
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %v", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	return token.AccessToken, nil
}
