package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// UploadFile uploads a file to local storage and returns the public URL
func UploadFile(file multipart.File, filename string, folder string) (string, error) {
	// Ensure the upload directory exists
	baseDir := "public/uploads"
	targetDir := filepath.Join(baseDir, folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Generate a unique filename
	ext := filepath.Ext(filename)
	name := filename[0 : len(filename)-len(ext)]
	newFilename := fmt.Sprintf("%s_%d%s", name, time.Now().UnixNano(), ext)
	path := filepath.Join(targetDir, newFilename)

	// Create the file
	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy the uploaded content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return the URL (assuming the server serves 'public/uploads' at '/uploads')
	return fmt.Sprintf("/uploads/%s/%s", folder, newFilename), nil
}
