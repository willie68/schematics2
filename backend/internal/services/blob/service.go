package blob

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
)

// Service writes binary payloads into rotating container files (*.cnt).
type Service struct {
	cfg          config.Repository
	log          *slog.Logger
	dir          string
	maxSizeBytes int64

	mu          sync.Mutex
	currentFile *os.File
	currentNum  int
	currentSize int64
}

func New(inj do.Injector) *Service {
	cfg := do.MustInvoke[config.Config](inj)

	return &Service{
		cfg: cfg.Repository,
		log: logging.New("blob-service"),
	}
}

func (s *Service) Prepare() error {
	if s.cfg.RepositoryPath == "" {
		return errors.New("repository path is empty")
	}
	s.log.Info("preparing blob service", "dir", s.cfg.RepositoryPath, "maxSizeMB", s.cfg.ContainerMaxSizeMB)

	s.dir = s.cfg.RepositoryPath
	s.maxSizeBytes = s.cfg.ContainerMaxSizeMB * 1024 * 1024
	if s.maxSizeBytes <= 0 {
		s.maxSizeBytes = 100 * 1024 * 1024
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	nums, err := s.listContainerNumbers()
	if err != nil {
		return err
	}

	if len(nums) == 0 {
		s.currentNum = 1
		return s.createNewContainer()
	}

	s.currentNum = nums[len(nums)-1]
	fname := filepath.Join(s.dir, fmt.Sprintf("%d.cnt", s.currentNum))
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}

	off, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		_ = f.Close()
		return err
	}

	s.currentFile = f
	s.currentSize = off
	if s.currentSize >= s.maxSizeBytes {
		_ = f.Close()
		s.currentNum++
		return s.createNewContainer()
	}

	return nil
}

func (s *Service) Save(data []byte, mimeType string) (*domain.ContainerInfo, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentFile == nil {
		if err := s.createNewContainer(); err != nil {
			return nil, err
		}
	}

	compressed, compressionType, err := s.compressData(data)
	if err != nil {
		return nil, fmt.Errorf("compress data: %w", err)
	}

	// Format: [4-byte original-length][1-byte compression-type][variable-length data]
	recordSize := int64(4 + 1 + len(compressed))
	if s.currentSize+recordSize > s.maxSizeBytes {
		_ = s.currentFile.Close()
		s.currentNum++
		if err := s.createNewContainer(); err != nil {
			return nil, err
		}
	}

	offset, err := s.currentFile.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	// Write original length
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err = s.currentFile.Write(lenBuf[:]); err != nil {
		return nil, err
	}

	// Write compression type
	if _, err = s.currentFile.Write([]byte{compressionType}); err != nil {
		return nil, err
	}

	// Write compressed data
	n, err := s.currentFile.Write(compressed)
	if err != nil {
		return nil, err
	}

	if err = s.currentFile.Sync(); err != nil {
		return nil, err
	}

	s.currentSize += recordSize

	ci := &domain.ContainerInfo{
		ContainerNumber: s.currentNum,
		Offset:          offset,
		Length:          int64(n + 5), // 4 bytes length + 1 byte compression type + data
		OriginalLength:  int64(len(data)),
		MIMEType:        mimeType,
		Compressed:      compressionTypeToString(compressionType),
	}

	// Persist container info to .inf file
	if err = s.appendContainerInfoEntry(s.currentNum, ci); err != nil {
		return nil, fmt.Errorf("persist container info: %w", err)
	}

	return ci, nil
}

func (s *Service) Load(ci *domain.ContainerInfo) ([]byte, error) {
	if ci == nil {
		return nil, errors.New("container info is nil")
	}

	fname := filepath.Join(s.dir, fmt.Sprintf("%d.cnt", ci.ContainerNumber))
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err = f.Seek(ci.Offset, io.SeekStart); err != nil {
		return nil, err
	}

	// Read original length
	var lenBuf [4]byte
	if _, err = io.ReadFull(f, lenBuf[:]); err != nil {
		return nil, err
	}
	originalLen := int64(binary.BigEndian.Uint32(lenBuf[:]))

	// Read compression type
	var compressionBuf [1]byte
	if _, err = io.ReadFull(f, compressionBuf[:]); err != nil {
		return nil, err
	}
	compressionType := compressionBuf[0]

	// Read compressed data
	// Actual compressed size is Length - 5 (4 bytes for original length + 1 byte for compression type)
	compressedSize := ci.Length - 5
	compressed := make([]byte, compressedSize)
	if _, err = io.ReadFull(f, compressed); err != nil {
		return nil, err
	}

	// Decompress if needed
	return s.decompressData(compressed, compressionType, int(originalLen))
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentFile != nil {
		err := s.currentFile.Close()
		s.currentFile = nil
		return err
	}
	return nil
}

func (s *Service) Health(ctx context.Context) error {
	s.mu.Lock()
	dir := s.dir
	cur := s.currentNum
	s.mu.Unlock()

	if dir == "" {
		return errors.New("repository path is empty")
	}

	if _, err := os.Stat(dir); err != nil {
		return err
	}

	fname := filepath.Join(dir, fmt.Sprintf("%d.cnt", cur))
	if f, err := os.OpenFile(fname, os.O_RDONLY, 0o644); err == nil {
		_ = f.Close()
	}

	tmp, err := os.CreateTemp(dir, "health-*.tmp")
	if err != nil {
		return err
	}
	if _, err = tmp.Write([]byte{0}); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	tmpName := tmp.Name()
	_ = tmp.Close()

	if _, err = os.ReadFile(tmpName); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	_ = os.Remove(tmpName)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (s *Service) createNewContainer() error {
	fname := filepath.Join(s.dir, fmt.Sprintf("%d.cnt", s.currentNum))
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	s.currentFile = f
	s.currentSize = 0
	return nil
}

func (s *Service) listContainerNumbers() ([]int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	nums := make([]int, 0)
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".cnt") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".cnt")
		n, convErr := strconv.Atoi(base)
		if convErr != nil {
			continue
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)

	return nums, nil
}

// Compression type constants
const (
	compressionNone byte = 0x00
	compressionZstd byte = 0x01
	compressionGzip byte = 0x02
)

func (s *Service) compressData(data []byte) ([]byte, byte, error) {
	compressionType := s.cfg.CompressionType
	if compressionType == "" {
		compressionType = "none"
	}

	switch strings.ToLower(compressionType) {
	case "none":
		return data, compressionNone, nil

	case "zstd":
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			return nil, 0, fmt.Errorf("create zstd encoder: %w", err)
		}
		defer encoder.Close()

		compressed := encoder.EncodeAll(data, nil)
		return compressed, compressionZstd, nil

	case "gzip":
		buf := bytes.Buffer{}
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			_ = writer.Close()
			return nil, 0, fmt.Errorf("gzip encode: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, 0, fmt.Errorf("gzip close: %w", err)
		}
		return buf.Bytes(), compressionGzip, nil

	default:
		return nil, 0, fmt.Errorf("unknown compression type: %q", compressionType)
	}
}

func (s *Service) decompressData(compressed []byte, compressionType byte, expectedLen int) ([]byte, error) {
	switch compressionType {
	case compressionNone:
		return compressed, nil

	case compressionZstd:
		decoder, err := zstd.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("create zstd decoder: %w", err)
		}
		defer decoder.Close()

		decompressed, err := io.ReadAll(decoder)
		if err != nil {
			return nil, fmt.Errorf("zstd decode: %w", err)
		}

		if len(decompressed) != expectedLen {
			return nil, fmt.Errorf("decompressed size mismatch: got %d, expected %d", len(decompressed), expectedLen)
		}

		return decompressed, nil

	case compressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("create gzip reader: %w", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip decode: %w", err)
		}

		if len(decompressed) != expectedLen {
			return nil, fmt.Errorf("decompressed size mismatch: got %d, expected %d", len(decompressed), expectedLen)
		}

		return decompressed, nil

	default:
		return nil, fmt.Errorf("unknown compression type: 0x%02x", compressionType)
	}
}

func compressionTypeToString(ct byte) string {
	switch ct {
	case compressionNone:
		return "none"
	case compressionZstd:
		return "zstd"
	case compressionGzip:
		return "gzip"
	default:
		return "unknown"
	}
}

// containerInfoEntry is the JSON-serializable format for .inf files
type containerInfoEntry struct {
	Offset         int64  `json:"offset"`
	Length         int64  `json:"length"`
	OriginalLength int64  `json:"originalLength"`
	MIMEType       string `json:"mimeType,omitempty"`
	Compressed     string `json:"compressed,omitempty"`
}

// appendContainerInfoEntry writes a ContainerInfo entry to the .inf file
func (s *Service) appendContainerInfoEntry(containerNum int, ci *domain.ContainerInfo) error {
	infPath := filepath.Join(s.dir, fmt.Sprintf("%d.inf", containerNum))

	// Load existing entries
	var entries []containerInfoEntry
	if data, err := os.ReadFile(infPath); err == nil {
		if err = json.Unmarshal(data, &entries); err != nil {
			s.log.Warn("failed to parse existing .inf file, starting fresh", "path", infPath, "error", err)
			entries = nil
		}
	}

	// Append new entry
	entries = append(entries, containerInfoEntry{
		Offset:         ci.Offset,
		Length:         ci.Length,
		OriginalLength: ci.OriginalLength,
		MIMEType:       ci.MIMEType,
		Compressed:     ci.Compressed,
	})

	// Write back to file
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal .inf data: %w", err)
	}

	if err = os.WriteFile(infPath, data, 0o644); err != nil {
		return fmt.Errorf("write .inf file: %w", err)
	}

	return nil
}

// LoadContainerInfos loads all ContainerInfo entries for a specific container from its .inf file
func (s *Service) LoadContainerInfos(containerNum int) ([]*domain.ContainerInfo, error) {
	infPath := filepath.Join(s.dir, fmt.Sprintf("%d.inf", containerNum))

	data, err := os.ReadFile(infPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Container exists but has no .inf file (pre-compression containers)
		}
		return nil, fmt.Errorf("read .inf file: %w", err)
	}

	var entries []containerInfoEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal .inf data: %w", err)
	}

	result := make([]*domain.ContainerInfo, len(entries))
	for i, entry := range entries {
		result[i] = &domain.ContainerInfo{
			ContainerNumber: containerNum,
			Offset:          entry.Offset,
			Length:          entry.Length,
			OriginalLength:  entry.OriginalLength,
			MIMEType:        entry.MIMEType,
			Compressed:      entry.Compressed,
		}
	}

	return result, nil
}

// ListAllContainerInfos iterates over all containers and returns all ContainerInfo entries
func (s *Service) ListAllContainerInfos() (map[int][]*domain.ContainerInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nums, err := s.listContainerNumbers()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make(map[int][]*domain.ContainerInfo)
	for _, num := range nums {
		infos, err := s.LoadContainerInfos(num)
		if err != nil {
			return nil, fmt.Errorf("load container infos for %d: %w", num, err)
		}
		if infos != nil {
			result[num] = infos
		}
	}

	return result, nil
}
