package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/tiff"

	"github.com/willie68/schematics2/backend/internal/domain/model"
)

// testLogger creates a logger for tests
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

// TestDownloadFile_TiffConversion tests the downloadFile endpoint with TIFF to PNG conversion
func TestDownloadFile_TiffConversion(t *testing.T) {
	// Create a test TIFF image
	rect := image.Rect(0, 0, 10, 10)
	img := image.NewRGBA(rect)
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 25), uint8(y * 25), 128, 255})
		}
	}

	var tiffBuf bytes.Buffer
	err := tiff.Encode(&tiffBuf, img, &tiff.Options{Compression: tiff.Uncompressed})
	require.NoError(t, err, "failed to create test TIFF")

	tiffData := tiffBuf.Bytes()
	encodedTiff := base64.StdEncoding.EncodeToString(tiffData)

	// Test WITH format=png
	t.Run("with_format_png", func(t *testing.T) {
		mockDocStore := newMockdocumentStore(t)
		mockBlobStore := newMockblobStore(t)

		testDoc := model.Document{
			ID:           "test-doc-1",
			Manufacturer: "TestMfg",
			Model:        "TestModel",
			Files: []model.DocumentFile{
				{
					Name:      "test.tif",
					Type:      "schematic",
					MIMEType:  "image/tiff",
					Page:      1,
					Container: &model.ContainerInfo{ID: "container-1"},
				},
			},
		}

		mockDocStore.EXPECT().GetByID(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), "test-doc-1").Return(testDoc, nil)
		mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "container-1"}).Return(tiffData, nil)

		h := &Handler{
			docStore: mockDocStore,
			blob:     mockBlobStore,
			log:      testLogger(),
		}

		req, err := http.NewRequest("GET", "/api/v1/documents/test-doc-1/files/test.tif?format=png", nil)
		require.NoError(t, err)
		req = req.WithContext(context.Background())

		// Add URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test-doc-1")
		rctx.URLParams.Add("filename", "test.tif")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.downloadFile(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "should return 200 OK")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "response should be valid JSON")

		// Check response fields
		assert.Equal(t, "test.tif", response["name"])
		assert.Equal(t, "schematic", response["type"])
		assert.Equal(t, "image/png", response["mimetype"], "should return PNG mimetype when converted")
		assert.NotEmpty(t, response["data"], "should have base64 data")

		// Decode and verify it's valid PNG
		pngBase64 := response["data"].(string)
		pngData, err := base64.StdEncoding.DecodeString(pngBase64)
		require.NoError(t, err, "should be valid base64")

		cfg, err := png.DecodeConfig(bytes.NewReader(pngData))
		require.NoError(t, err, "should be valid PNG")
		assert.Equal(t, 10, cfg.Width)
		assert.Equal(t, 10, cfg.Height)
	})

	// Test WITHOUT format parameter (returns original TIFF)
	t.Run("without_format_param", func(t *testing.T) {
		mockDocStore := newMockdocumentStore(t)
		mockBlobStore := newMockblobStore(t)

		testDoc := model.Document{
			ID:           "test-doc-1",
			Manufacturer: "TestMfg",
			Model:        "TestModel",
			Files: []model.DocumentFile{
				{
					Name:      "test.tif",
					Type:      "schematic",
					MIMEType:  "image/tiff",
					Page:      1,
					Container: &model.ContainerInfo{ID: "container-1"},
				},
			},
		}

		mockDocStore.EXPECT().GetByID(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), "test-doc-1").Return(testDoc, nil)
		mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "container-1"}).Return(tiffData, nil)

		h := &Handler{
			docStore: mockDocStore,
			blob:     mockBlobStore,
			log:      testLogger(),
		}

		req, err := http.NewRequest("GET", "/api/v1/documents/test-doc-1/files/test.tif", nil)
		require.NoError(t, err)
		req = req.WithContext(context.Background())

		// Add URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test-doc-1")
		rctx.URLParams.Add("filename", "test.tif")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.downloadFile(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "should return 200 OK")

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "response should be valid JSON")

		// Should return original TIFF mimetype
		assert.Equal(t, "image/tiff", response["mimetype"], "should return original TIFF mimetype")

		// Verify data is original TIFF
		tiffBase64 := response["data"].(string)
		assert.Equal(t, encodedTiff, tiffBase64, "data should be original TIFF")
	})
}

func TestDownloadFile_TiffConversion_AllMimeTypes(t *testing.T) {
	// Create test TIFF data
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	var tiffBuf bytes.Buffer
	err := tiff.Encode(&tiffBuf, img, &tiff.Options{Compression: tiff.Uncompressed})
	require.NoError(t, err)

	tiffData := tiffBuf.Bytes()

	testCases := []struct {
		mimeType string
		desc     string
	}{
		{"image/tiff", "standard TIFF"},
		{"image/x-tiff", "x-tiff variant"},
		{"image/vnd.tiff", "vendor TIFF"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockDocStore := newMockdocumentStore(t)
			mockBlobStore := newMockblobStore(t)

			testDoc := model.Document{
				ID: "test-doc",
				Files: []model.DocumentFile{
					{
						Name:      "test.tif",
						Type:      "schematic",
						MIMEType:  tc.mimeType,
						Container: &model.ContainerInfo{ID: "container-1"},
					},
				},
			}

			mockDocStore.EXPECT().GetByID(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), "test-doc").Return(testDoc, nil)
			mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "container-1"}).Return(tiffData, nil)

			h := &Handler{
				docStore: mockDocStore,
				blob:     mockBlobStore,
				log:      testLogger(),
			}

			req, err := http.NewRequest("GET", "/api/v1/documents/test-doc/files/test.tif?format=png", nil)
			require.NoError(t, err)
			req = req.WithContext(context.Background())

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "test-doc")
			rctx.URLParams.Add("filename", "test.tif")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			h.downloadFile(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			// Should be converted to PNG regardless of TIFF variant
			assert.Equal(t, "image/png", response["mimetype"], "should convert %s to PNG", tc.mimeType)
		})
	}
}

func TestDownloadFile_NonTiffNotConverted(t *testing.T) {
	// Create a PNG image
	rect := image.Rect(0, 0, 2, 2)
	img := image.NewRGBA(rect)
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	var pngBuf bytes.Buffer
	err := png.Encode(&pngBuf, img)
	require.NoError(t, err)

	pngData := pngBuf.Bytes()
	encodedPng := base64.StdEncoding.EncodeToString(pngData)

	mockDocStore := newMockdocumentStore(t)
	mockBlobStore := newMockblobStore(t)

	testDoc := model.Document{
		ID: "test-doc",
		Files: []model.DocumentFile{
			{
				Name:      "test.png",
				Type:      "schematic",
				MIMEType:  "image/png",
				Container: &model.ContainerInfo{ID: "container-1"},
			},
		},
	}

	mockDocStore.EXPECT().GetByID(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), "test-doc").Return(testDoc, nil)
	mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "container-1"}).Return(pngData, nil)

	h := &Handler{
		docStore: mockDocStore,
		blob:     mockBlobStore,
		log:      testLogger(),
	}

	req, err := http.NewRequest("GET", "/api/v1/documents/test-doc/files/test.png?format=png", nil)
	require.NoError(t, err)
	req = req.WithContext(context.Background())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-doc")
	rctx.URLParams.Add("filename", "test.png")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.downloadFile(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// PNG should NOT be converted
	assert.Equal(t, "image/png", response["mimetype"], "PNG should remain PNG")
	assert.Equal(t, encodedPng, response["data"], "PNG data should be unchanged")
}

func TestDownloadFile_TiffConversionError(t *testing.T) {
	// Create invalid TIFF data
	invalidTiffData := []byte{0x49, 0x49, 0x2A, 0x00} // TIFF header only

	mockDocStore := newMockdocumentStore(t)
	mockBlobStore := newMockblobStore(t)

	testDoc := model.Document{
		ID: "test-doc",
		Files: []model.DocumentFile{
			{
				Name:      "test.tif",
				Type:      "schematic",
				MIMEType:  "image/tiff",
				Container: &model.ContainerInfo{ID: "container-1"},
			},
		},
	}

	mockDocStore.EXPECT().GetByID(mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }), "test-doc").Return(testDoc, nil)
	mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "container-1"}).Return(invalidTiffData, nil)

	h := &Handler{
		docStore: mockDocStore,
		blob:     mockBlobStore,
		log:      testLogger(),
	}

	req, err := http.NewRequest("GET", "/api/v1/documents/test-doc/files/test.tif?format=png", nil)
	require.NoError(t, err)
	req = req.WithContext(context.Background())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-doc")
	rctx.URLParams.Add("filename", "test.tif")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.downloadFile(w, req)

	// When conversion fails, should still return original TIFF
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should fall back to original TIFF mimetype
	assert.Equal(t, "image/tiff", response["mimetype"], "should fall back to TIFF on conversion error")
}
