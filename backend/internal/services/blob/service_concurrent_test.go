package blob

import (
	"bytes"
	"sync"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
)

type BlobServiceConcurrentTestSuite struct {
	suite.Suite
	inj     do.Injector
	tempDir string
	svc     *Service
}

func (s *BlobServiceConcurrentTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
	s.inj = do.New()

	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1, // Small to force rotation
		},
	}
	do.ProvideValue(s.inj, cfg)

	s.svc = New(s.inj)
	s.Require().NotNil(s.svc)
}

func (s *BlobServiceConcurrentTestSuite) TearDownTest() {
	if s.svc != nil {
		_ = s.svc.Close()
	}
}

// TestConcurrentSavesToDifferentContainers verifies that sequential saves to different containers
// via rotation don't corrupt data when accessed concurrently
func (s *BlobServiceConcurrentTestSuite) TestConcurrentSavesToDifferentContainers() {
	// GIVEN - prepare service
	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	// Sequential saves to force rotation
	payload1 := bytes.Repeat([]byte("A"), 600*1024) // 600KB - Container 1, triggers rotation
	payload2 := bytes.Repeat([]byte("B"), 600*1024) // 600KB - Container 2
	payload3 := bytes.Repeat([]byte("C"), 400*1024) // 400KB - Container 2

	// Save sequentially to different containers
	ci1, err := s.svc.Save(payload1, "application/pdf")
	s.Require().NoError(err, "save first payload")

	ci2, err := s.svc.Save(payload2, "application/pdf")
	s.Require().NoError(err, "save second payload")

	ci3, err := s.svc.Save(payload3, "application/pdf")
	s.Require().NoError(err, "save third payload")

	// WHEN - load concurrently from multiple goroutines
	var wg sync.WaitGroup
	var err1, err2, err3 error
	var actualData1, actualData2, actualData3 []byte
	var mu sync.Mutex

	wg.Add(3)

	// Goroutine 1: Load from container 1
	go func() {
		defer wg.Done()
		d, e := s.svc.Load(ci1)
		mu.Lock()
		actualData1, err1 = d, e
		mu.Unlock()
	}()

	// Goroutine 2: Load from container 2
	go func() {
		defer wg.Done()
		d, e := s.svc.Load(ci2)
		mu.Lock()
		actualData2, err2 = d, e
		mu.Unlock()
	}()

	// Goroutine 3: Load from container 2
	go func() {
		defer wg.Done()
		d, e := s.svc.Load(ci3)
		mu.Lock()
		actualData3, err3 = d, e
		mu.Unlock()
	}()

	wg.Wait()

	// THEN - verify all loads succeeded
	s.Require().NoError(err1, "first concurrent load should succeed")
	s.Require().NoError(err2, "second concurrent load should succeed")
	s.Require().NoError(err3, "third concurrent load should succeed")

	s.Assert().Equal(payload1, actualData1, "first payload should match")
	s.Assert().Equal(payload2, actualData2, "second payload should match")
	s.Assert().Equal(payload3, actualData3, "third payload should match")

	// Verify that containers have expected structure
	s.Assert().Equal(1, ci1.ContainerNumber, "first save should be to container 1")
	s.Assert().Equal(2, ci2.ContainerNumber, "second save should be to container 2 due to rotation")
	s.Assert().Equal(2, ci3.ContainerNumber, "third save should be to container 2")

	// Verify .inf files
	infos1, err := s.svc.LoadContainerInfos(1)
	s.Require().NoError(err, "load container 1 infos")
	s.Require().Len(infos1, 1, "container 1 should have 1 entry")

	infos2, err := s.svc.LoadContainerInfos(2)
	s.Require().NoError(err, "load container 2 infos")
	s.Require().Len(infos2, 2, "container 2 should have 2 entries")
}

// TestConcurrentSavesToSameContainer verifies serialization when multiple goroutines
// try to write to the same container (they should be serialized by container lock)
func (s *BlobServiceConcurrentTestSuite) TestConcurrentSavesToSameContainer() {
	// GIVEN - prepare service with large container to prevent rotation
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 100, // Large - prevents rotation
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service with large container")

	// Small payloads that fit in one container
	payload1 := []byte("payload-1")
	payload2 := []byte("payload-2")
	payload3 := []byte("payload-3")

	// WHEN - save concurrently from multiple goroutines to same container
	var wg sync.WaitGroup
	results := make([]*domain.ContainerInfo, 3)
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		idx := i
		go func(payload []byte) {
			defer wg.Done()
			ci, e := s.svc.Save(payload, "text/plain")
			s.Require().NoError(e, "concurrent save should succeed")
			mu.Lock()
			results[idx] = ci
			mu.Unlock()
		}([][]byte{payload1, payload2, payload3}[i])
	}

	wg.Wait()

	// THEN - verify all saves went to same container (no rotation)
	s.Assert().Equal(1, results[0].ContainerNumber, "all saves should be to container 1")
	s.Assert().Equal(1, results[1].ContainerNumber, "all saves should be to container 1")
	s.Assert().Equal(1, results[2].ContainerNumber, "all saves should be to container 1")

	// Verify all entries are in .inf file
	infos, err := s.svc.LoadContainerInfos(1)
	s.Require().NoError(err, "load container infos")
	s.Require().Len(infos, 3, "container should have 3 entries")

	// Verify all can be loaded
	data1, _ := s.svc.Load(results[0])
	data2, _ := s.svc.Load(results[1])
	data3, _ := s.svc.Load(results[2])

	s.Assert().Equal(payload1, data1, "first payload should match")
	s.Assert().Equal(payload2, data2, "second payload should match")
	s.Assert().Equal(payload3, data3, "third payload should match")
}

func TestBlobServiceConcurrentTestSuite(t *testing.T) {
	suite.Run(t, new(BlobServiceConcurrentTestSuite))
}
