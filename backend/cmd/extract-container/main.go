package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

// containerInfoEntry represents a single entry in the .inf file
type containerInfoEntry struct {
	Offset         int64  `json:"offset"`
	Length         int64  `json:"length"`
	OriginalLength int64  `json:"originalLength"`
	MIMEType       string `json:"mimeType,omitempty"`
	Compressed     string `json:"compressed,omitempty"`
}

const (
	compressionNone byte = 0x00
	compressionZstd byte = 0x01
	compressionGzip byte = 0x02
)

func main() {
	repoDir := flag.String("repo", "./dev_data/repository", "Repository directory containing .cnt and .inf files")
	containerNum := flag.Int("container", 0, "Container number (required)")
	offset := flag.Int64("offset", 0, "Byte offset in container (required)")
	outFile := flag.String("out", "", "Output file path (default: auto-generated from MIME type)")

	flag.Parse()

	// Validate inputs
	if *containerNum <= 0 {
		log.Fatal("Error: -container must be > 0")
	}
	if *offset < 0 {
		log.Fatal("Error: -offset must be >= 0")
	}

	// Read .inf file to find the entry
	infPath := filepath.Join(*repoDir, fmt.Sprintf("%d.inf", *containerNum))
	infData, err := os.ReadFile(infPath)
	if err != nil {
		log.Fatalf("Failed to read .inf file: %v", err)
	}

	var entries []containerInfoEntry
	if err := json.Unmarshal(infData, &entries); err != nil {
		log.Fatalf("Failed to parse .inf file: %v", err)
	}

	// Find the entry at the given offset
	var entry *containerInfoEntry
	for i := range entries {
		if entries[i].Offset == *offset {
			entry = &entries[i]
			break
		}
	}

	if entry == nil {
		log.Fatalf("No entry found at offset %d in container %d", *offset, *containerNum)
	}

	fmt.Printf("Found entry at offset %d:\n", *offset)
	fmt.Printf("  Length: %d bytes (compressed)\n", entry.Length)
	fmt.Printf("  Original Length: %d bytes\n", entry.OriginalLength)
	fmt.Printf("  MIME Type: %s\n", entry.MIMEType)
	fmt.Printf("  Compression: %s\n", entry.Compressed)

	// Open and read the container file
	cntPath := filepath.Join(*repoDir, fmt.Sprintf("%d.cnt", *containerNum))
	cntFile, err := os.Open(cntPath)
	if err != nil {
		log.Fatalf("Failed to open container file: %v", err)
	}
	defer cntFile.Close()

	// Seek to the offset
	if _, err := cntFile.Seek(*offset, io.SeekStart); err != nil {
		log.Fatalf("Failed to seek in container file: %v", err)
	}

	// Read the data (with header: 4-byte original length + 1-byte compression type + data)
	header := make([]byte, 5)
	if _, err := io.ReadFull(cntFile, header); err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	originalLen := int64(binary.BigEndian.Uint32(header[:4]))
	compressionByte := header[4]

	fmt.Printf("  Compression type byte: 0x%02x\n", compressionByte)

	// Read compressed data
	dataSize := entry.Length - 5 // Total size minus header
	compressed := make([]byte, dataSize)
	if _, err := io.ReadFull(cntFile, compressed); err != nil {
		log.Fatalf("Failed to read data: %v", err)
	}

	// Decompress if needed
	var data []byte
	switch entry.Compressed {
	case "none":
		data = compressed
	case "zstd":
		decoder, err := zstd.NewReader(bytes.NewReader(compressed))
		if err != nil {
			log.Fatalf("Failed to create zstd decoder: %v", err)
		}
		defer decoder.Close()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, decoder); err != nil {
			log.Fatalf("Failed to decompress zstd: %v", err)
		}
		data = buf.Bytes()
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			log.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.Close()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, reader); err != nil {
			log.Fatalf("Failed to decompress gzip: %v", err)
		}
		data = buf.Bytes()
	default:
		log.Fatalf("Unknown compression type: %s", entry.Compressed)
	}

	// Verify decompressed size
	if int64(len(data)) != originalLen {
		log.Printf("Warning: Decompressed size mismatch: got %d, expected %d", len(data), originalLen)
	}

	// Generate output filename if not provided
	if *outFile == "" {
		ext := mimeToExt(entry.MIMEType)
		*outFile = fmt.Sprintf("output_%d_%d%s", *containerNum, *offset, ext)
	}

	// Write to output file
	if err := os.WriteFile(*outFile, data, 0o644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("✓ Extracted %d bytes to %s\n", len(data), *outFile)
}

// mimeToExt converts MIME type to file extension
func mimeToExt(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/x-zip-compressed":
		return ".zip"
	case "application/octet-stream":
		return ".bin"
	case "text/html; charset=utf-8":
		return ".html"
	default:
		return ".dat"
	}
}
