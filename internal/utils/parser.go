package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

// ReadJSON reads the body from an HTTP request and decodes it into the provided interface.
// This is a Go equivalent to the previous uWebSockets.js readJson helper, adapted for net/http.
func ReadJSON(r *http.Request, v interface{}) error {
	// Ensure the body is closed after reading
	defer r.Body.Close()

	// Read all data from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the target interface
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}
