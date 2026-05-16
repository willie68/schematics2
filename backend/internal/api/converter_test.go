package api

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/tiff"
)

// createTestTiffImage creates a small 2x2 test TIFF image
func createTestTiffImage(t *testing.T) []byte {
	// Create a simple 2x2 image
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)

	// Set some colors
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})   // Red
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})   // Green
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})   // Blue
	img.Set(1, 1, color.RGBA{255, 255, 0, 255}) // Yellow

	// Encode as TIFF
	var buf bytes.Buffer
	err := tiff.Encode(&buf, img, &tiff.Options{Compression: tiff.Uncompressed})
	require.NoError(t, err, "failed to create test TIFF image")

	return buf.Bytes()
}

// createTestPngImage creates a small test PNG image
func createTestPngImage(t *testing.T) []byte {
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 1, color.RGBA{0, 0, 255, 255})

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err, "failed to create test PNG image")

	return buf.Bytes()
}

// createTestJpegImage creates a small test JPEG image
func createTestJpegImage(t *testing.T) []byte {
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 1, color.RGBA{0, 0, 255, 255})

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err, "failed to create test JPEG image")

	return buf.Bytes()
}

func TestConvertTiffToPng_Success(t *testing.T) {
	tiffData := createTestTiffImage(t)
	require.NotEmpty(t, tiffData, "test TIFF image should not be empty")

	pngData, err := convertTiffToPng(tiffData)

	assert.NoError(t, err, "convertTiffToPng should not return error for valid TIFF")
	assert.NotEmpty(t, pngData, "converted PNG data should not be empty")

	// Verify the result is valid PNG by decoding it
	img, err := png.DecodeConfig(bytes.NewReader(pngData))
	assert.NoError(t, err, "result should be valid PNG")
	assert.Equal(t, 2, img.Width, "PNG width should match original")
	assert.Equal(t, 2, img.Height, "PNG height should match original")
}

func TestConvertTiffToPng_InvalidTiff(t *testing.T) {
	// Use PNG data instead of TIFF - should fail
	pngData := createTestPngImage(t)

	_, err := convertTiffToPng(pngData)

	assert.Error(t, err, "convertTiffToPng should return error for non-TIFF data")
	assert.Contains(t, err.Error(), "decode tiff", "error should mention TIFF decode issue")
}

func TestConvertTiffToPng_EmptyData(t *testing.T) {
	_, err := convertTiffToPng([]byte{})

	assert.Error(t, err, "convertTiffToPng should return error for empty data")
	assert.Contains(t, err.Error(), "decode tiff", "error should mention TIFF decode issue")
}

func TestConvertTiffToPng_NilData(t *testing.T) {
	_, err := convertTiffToPng(nil)

	assert.Error(t, err, "convertTiffToPng should return error for nil data")
}

func TestConvertTiffToPng_CorruptedData(t *testing.T) {
	corruptedData := []byte{0x49, 0x49, 0x2A, 0x00} // TIFF header only, no image data

	_, err := convertTiffToPng(corruptedData)

	assert.Error(t, err, "convertTiffToPng should return error for corrupted TIFF")
	assert.Contains(t, err.Error(), "decode tiff", "error should mention decode issue")
}

func TestConvertTiffToPng_JpegData(t *testing.T) {
	jpegData := createTestJpegImage(t)

	_, err := convertTiffToPng(jpegData)

	assert.Error(t, err, "convertTiffToPng should return error for JPEG data")
	assert.Contains(t, err.Error(), "decode tiff", "error should mention TIFF decode issue")
}

func TestIsTiffMimeType_ValidTiffMimes(t *testing.T) {
	testCases := []struct {
		mimeType string
		expected bool
		desc     string
	}{
		{"image/tiff", true, "standard TIFF MIME type"},
		{"image/x-tiff", true, "alternative x-tiff MIME type"},
		{"image/vnd.tiff", true, "vendor-specific TIFF MIME type"},
		{"IMAGE/TIFF", true, "uppercase should be recognized"},
		{"Image/Tiff", true, "mixed case should be recognized"},
		{"  image/tiff  ", true, "whitespace should be trimmed"},
		{"\timage/tiff\n", true, "tabs and newlines should be trimmed"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isTiffMimeType(tc.mimeType)
			assert.Equal(t, tc.expected, result, "isTiffMimeType failed for: %s", tc.mimeType)
		})
	}
}

func TestIsTiffMimeType_InvalidMimes(t *testing.T) {
	testCases := []struct {
		mimeType string
		desc     string
	}{
		{"image/png", "PNG MIME type"},
		{"image/jpeg", "JPEG MIME type"},
		{"image/gif", "GIF MIME type"},
		{"image/bmp", "BMP MIME type"},
		{"image/webp", "WebP MIME type"},
		{"application/pdf", "PDF MIME type"},
		{"", "empty string"},
		{"   ", "only whitespace"},
		{"not/amime", "invalid format"},
		{"tiff", "without image/ prefix"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isTiffMimeType(tc.mimeType)
			assert.False(t, result, "isTiffMimeType should return false for: %s", tc.mimeType)
		})
	}
}

func TestConvertTiffToPng_DifferentSizes(t *testing.T) {
	// Test with larger images
	sizes := []struct {
		width  int
		height int
		desc   string
	}{
		{1, 1, "1x1 pixel"},
		{10, 10, "10x10 pixels"},
		{100, 100, "100x100 pixels"},
	}

	for _, size := range sizes {
		t.Run(size.desc, func(t *testing.T) {
			// Create test TIFF
			rect := image.Rect(0, 0, size.width, size.height)
			img := image.NewRGBA(rect)
			for x := 0; x < size.width; x++ {
				for y := 0; y < size.height; y++ {
					img.Set(x, y, color.RGBA{byte(x % 256), byte(y % 256), 128, 255})
				}
			}

			var tiffBuf bytes.Buffer
			err := tiff.Encode(&tiffBuf, img, &tiff.Options{Compression: tiff.Uncompressed})
			require.NoError(t, err, "failed to create TIFF for size test")

			// Convert
			pngData, err := convertTiffToPng(tiffBuf.Bytes())
			assert.NoError(t, err, "conversion should succeed for %s", size.desc)
			assert.NotEmpty(t, pngData, "PNG data should not be empty for %s", size.desc)

			// Verify dimensions
			config, err := png.DecodeConfig(bytes.NewReader(pngData))
			assert.NoError(t, err, "result should be valid PNG for %s", size.desc)
			assert.Equal(t, size.width, config.Width, "PNG width mismatch for %s", size.desc)
			assert.Equal(t, size.height, config.Height, "PNG height mismatch for %s", size.desc)
		})
	}
}

func TestConvertTiffToPng_ColorPreservation(t *testing.T) {
	// Create test image with specific colors
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)

	testColors := []struct {
		x, y  int
		color color.RGBA
	}{
		{0, 0, color.RGBA{255, 0, 0, 255}},     // Red
		{1, 0, color.RGBA{0, 255, 0, 255}},     // Green
		{0, 1, color.RGBA{0, 0, 255, 255}},     // Blue
		{1, 1, color.RGBA{255, 255, 255, 255}}, // White
	}

	for _, tc := range testColors {
		img.Set(tc.x, tc.y, tc.color)
	}

	// Encode to TIFF
	var tiffBuf bytes.Buffer
	err := tiff.Encode(&tiffBuf, img, &tiff.Options{Compression: tiff.Uncompressed})
	require.NoError(t, err, "failed to encode TIFF")

	// Convert to PNG
	pngData, err := convertTiffToPng(tiffBuf.Bytes())
	require.NoError(t, err, "conversion should succeed")

	// Decode PNG and verify colors are preserved
	decodedImg, err := png.Decode(bytes.NewReader(pngData))
	require.NoError(t, err, "PNG decode should succeed")

	for _, tc := range testColors {
		r, g, b, a := decodedImg.At(tc.x, tc.y).RGBA()
		// RGBA returns 16-bit values, need to convert to 8-bit
		r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

		assert.Equal(t, tc.color.R, r8, "Red channel mismatch at (%d,%d)", tc.x, tc.y)
		assert.Equal(t, tc.color.G, g8, "Green channel mismatch at (%d,%d)", tc.x, tc.y)
		assert.Equal(t, tc.color.B, b8, "Blue channel mismatch at (%d,%d)", tc.x, tc.y)
		assert.Equal(t, tc.color.A, a8, "Alpha channel mismatch at (%d,%d)", tc.x, tc.y)
	}
}

func TestConvertTiffToPng_PerformanceReasonable(t *testing.T) {
	// Test that conversion doesn't take too long for reasonable image size
	rect := image.Rect(0, 0, 500, 500)
	img := image.NewRGBA(rect)

	// Fill with some pattern
	for x := 0; x < 500; x++ {
		for y := 0; y < 500; y++ {
			img.Set(x, y, color.RGBA{uint8((x + y) % 256), uint8(x % 256), uint8(y % 256), 255})
		}
	}

	var tiffBuf bytes.Buffer
	err := tiff.Encode(&tiffBuf, img, &tiff.Options{Compression: tiff.Uncompressed})
	require.NoError(t, err, "failed to encode TIFF")

	// This should complete quickly
	pngData, err := convertTiffToPng(tiffBuf.Bytes())
	assert.NoError(t, err, "conversion should succeed")
	assert.NotEmpty(t, pngData, "result should not be empty")
}
