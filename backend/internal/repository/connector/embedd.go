package connector

import (
	"embed"
	"io/fs"
)

// Note: This embed.FS is populated by the import-effecttypes tool
// which copies image files here during the import process.
// Image files (*.jpg, *.png, etc.) are created when running:
//
//	go run ./cmd/import-effecttypes/main.go
//
//go:embed *.png
var connectImages embed.FS

// GetImages returns the embedded filesystem containing all effect type images
func GetImages() embed.FS {
	return connectImages
}

// GetImage reads and returns a single image file by name
func GetImage(name string) ([]byte, error) {
	return fs.ReadFile(connectImages, name)
}

// ListImages returns all available image filenames
func ListImages() ([]string, error) {
	var images []string
	entries, err := fs.ReadDir(connectImages, ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			images = append(images, entry.Name())
		}
	}
	return images, nil
}
