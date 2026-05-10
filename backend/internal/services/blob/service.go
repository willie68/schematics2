package blob

import (
	"context"
	"encoding/binary"
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

	if s.currentSize+4+int64(len(data)) > s.maxSizeBytes {
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

	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err = s.currentFile.Write(lenBuf[:]); err != nil {
		return nil, err
	}

	n, err := s.currentFile.Write(data)
	if err != nil {
		return nil, err
	}

	if err = s.currentFile.Sync(); err != nil {
		return nil, err
	}

	s.currentSize += int64(n) + 4

	return &domain.ContainerInfo{
		ContainerNumber: s.currentNum,
		Offset:          offset,
		Length:          int64(n),
		MIMEType:        mimeType,
	}, nil
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

	var lenBuf [4]byte
	if _, err = io.ReadFull(f, lenBuf[:]); err != nil {
		return nil, err
	}

	dataLen := int64(binary.BigEndian.Uint32(lenBuf[:]))
	if dataLen <= 0 {
		return nil, errors.New("invalid data length in container")
	}

	buf := make([]byte, dataLen)
	if _, err = io.ReadFull(f, buf); err != nil {
		return nil, err
	}

	return buf, nil
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
