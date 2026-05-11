package effecttypes

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
//go:embed *.jpg
var effectTypeImages embed.FS

// GetImages returns the embedded filesystem containing all effect type images
func GetImages() embed.FS {
	return effectTypeImages
}

// GetImage reads and returns a single image file by name
func GetImage(filename string) ([]byte, error) {
	return fs.ReadFile(effectTypeImages, filename)
}

// ListImages returns all available image filenames
func ListImages() ([]string, error) {
	var images []string
	entries, err := fs.ReadDir(effectTypeImages, ".")
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
