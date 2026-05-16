package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type SearchResponse struct {
	Total int `json:"total"`
}

type Result struct {
	FilePath string
	FileName string
	Found    bool
	Error    error
}

func main() {
	rootDir := flag.String("dir", "", "Root directory to scan recursively")
	baseURL := flag.String("url", "", "Backend base URL (e.g., http://localhost:8080)")
	workers := flag.Int("workers", 5, "Number of concurrent workers for REST calls")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP request timeout")
	delete := flag.Bool("delete", false, "Delete found files from filesystem")

	flag.Parse()

	if *rootDir == "" || *baseURL == "" {
		fmt.Println("Usage:")
		fmt.Println("  check-missing-documents -dir=<ROOT_DIR> -url=<BASE_URL> [-workers=5] [-timeout=5s] [-delete]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println(`  check-missing-documents -dir="E:\Schematic_Files" -url="http://localhost:8080"`)
		fmt.Println()
		fmt.Println("Example with delete:")
		fmt.Println(`  check-missing-documents -dir="E:\Schematic_Files" -url="http://localhost:8080" -delete`)
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Verify directory exists
	if _, err := os.Stat(*rootDir); os.IsNotExist(err) {
		fmt.Printf("Error: Directory not found: %s\n", *rootDir)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("Document Presence Check Tool")
	fmt.Println("========================================")
	fmt.Printf("Root Directory: %s\n", *rootDir)
	fmt.Printf("Backend URL:    %s\n", *baseURL)
	fmt.Printf("Workers:        %d\n", *workers)
	fmt.Printf("Timeout:        %v\n", *timeout)
	if *delete {
		fmt.Println("Mode:           SEARCH + DELETE found files")
	} else {
		fmt.Println("Mode:           SEARCH ONLY")
	}
	fmt.Println("========================================")
	fmt.Println()

	// Collect all files
	var files []string
	err := filepath.Walk(*rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No files found in %s\n", *rootDir)
		os.Exit(0)
	}

	fmt.Printf("Found %d files to check...\n", len(files))
	fmt.Println()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: *timeout,
	}

	// Process files with worker pool
	results := processFilesWithWorkers(files, *baseURL, client, *workers)

	// Collect statistics
	var (
		found       int
		missing     int
		foundList   []string
		missingList []string
		errors      int
	)

	for result := range results {
		if result.Error != nil {
			errors++
			fmt.Printf("[ERROR] %s: %v\n", result.FileName, result.Error)
		} else if result.Found {
			found++
			foundList = append(foundList, result.FilePath)
		} else {
			missing++
			missingList = append(missingList, result.FilePath)
			fmt.Printf("[MISSING] %s\n", result.FileName)
		}

		// Progress indicator
		current := found + missing + errors
		if current%10 == 0 {
			fmt.Printf("Progress: %d/%d files checked, %d found, %d missing\n", current, len(files), found, missing)
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("Results")
	fmt.Println("========================================")
	fmt.Printf("Total Files:    %d\n", len(files))
	fmt.Printf("Found:          %d\n", found)
	fmt.Printf("Missing:        %d\n", missing)
	fmt.Printf("Errors:         %d\n", errors)
	fmt.Println("========================================")
	fmt.Println()

	// Write results to files
	if err := writeResultsToFiles(foundList, missingList); err != nil {
		fmt.Printf("Error writing results to files: %v\n", err)
		os.Exit(1)
	}

	if missing > 0 {
		fmt.Println("The following files were NOT found in the backend:")
		fmt.Println("(See missing.txt for complete list)")
		fmt.Println()
		// Show first 20
		for i, file := range missingList {
			if i >= 20 {
				fmt.Printf("... and %d more\n", len(missingList)-20)
				break
			}
			fmt.Println(file)
		}
		fmt.Println()
	} else {
		fmt.Println("All files were found in the backend! ✓")
	}

	fmt.Printf("Results written to:\n")
	fmt.Printf("  ✓ found.txt (%d files)\n", len(foundList))
	fmt.Printf("  ✓ missing.txt (%d files)\n", len(missingList))
}

func processFilesWithWorkers(files []string, baseURL string, client *http.Client, numWorkers int) <-chan Result {
	results := make(chan Result, len(files))
	var wg sync.WaitGroup

	// Create channel for work
	work := make(chan string, len(files))

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range work {
				fileName := filepath.Base(filePath)
				fileNameNoExt := stripExtension(fileName)

				found, err := searchBackend(baseURL, fileNameNoExt, client)
				results <- Result{
					FilePath: filePath,
					FileName: fileNameNoExt,
					Found:    found,
					Error:    err,
				}
			}
		}()
	}

	// Send work
	go func() {
		for _, file := range files {
			work <- file
		}
		close(work)
	}()

	// Wait for completion and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func stripExtension(fileName string) string {
	// Common extensions to remove
	extensions := []string{".tiff", ".tif", ".jpeg", ".jpg", ".png", ".gif", ".pdf", ".bmp"}

	lower := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lower, ext) {
			return fileName[:len(fileName)-len(ext)]
		}
	}

	// Fallback: remove last extension
	if idx := strings.LastIndex(fileName, "."); idx > 0 {
		return fileName[:idx]
	}

	return fileName
}

func searchBackend(baseURL, searchTerm string, client *http.Client) (bool, error) {
	searchURL := fmt.Sprintf("%s/api/v1/documents/search?q=%s", strings.TrimRight(baseURL, "/"), url.QueryEscape(searchTerm))

	resp, err := client.Get(searchURL)
	if err != nil {
		return false, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	return searchResp.Total > 0, nil
}

func writeResultsToFiles(foundList, missingList []string) error {
	// Write found.txt
	foundFile, err := os.Create("found.txt")
	if err != nil {
		return fmt.Errorf("failed to create found.txt: %w", err)
	}
	defer foundFile.Close()

	for _, file := range foundList {
		fmt.Fprintln(foundFile, file)
	}

	// Write missing.txt
	missingFile, err := os.Create("missing.txt")
	if err != nil {
		return fmt.Errorf("failed to create missing.txt: %w", err)
	}
	defer missingFile.Close()

	for _, file := range missingList {
		fmt.Fprintln(missingFile, file)
	}

	return nil
}
