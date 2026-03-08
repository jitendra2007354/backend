package services

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	path := filepath.Join("public/uploads", header.Filename)
	out, _ := os.Create(path)
	defer out.Close()
	io.Copy(out, file)
	return path, nil
}
