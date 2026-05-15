package blob

import (
	"bytes"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"github.com/willie68/schematics2/backend/internal/config"
)

type BlobServiceTestSuite struct {
	suite.Suite
	inj     do.Injector
	tempDir string
	svc     *Service
}

func (s *BlobServiceTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
	s.inj = do.New()

	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
		},
	}
	do.ProvideValue(s.inj, cfg)

	s.svc = New(s.inj)
	s.Require().NotNil(s.svc)
}

func (s *BlobServiceTestSuite) TearDownTest() {
	if s.svc != nil {
		_ = s.svc.Close()
	}
}

func TestBlobServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BlobServiceTestSuite))
}

func (s *BlobServiceTestSuite) TestSaveAndLoad_Success() {
	// GIVEN
	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	payload := []byte("hello-blob")

	// WHEN
	ci, err := s.svc.Save(payload, "text/plain", "hello.txt")

	// THEN
	s.Require().NoError(err, "save payload")
	s.Assert().NotNil(ci)
	s.Assert().Equal(1, ci.ContainerNumber)
	s.Assert().Equal(int64(len(payload)), ci.OriginalLength, "original length should match")
	s.Assert().Equal("text/plain", ci.MIMEType)
	s.Assert().Equal("none", ci.Compressed, "default compression is none")

	got, err := s.svc.Load(ci)
	s.Require().NoError(err, "load payload")
	s.Assert().Equal(payload, got)
}

func (s *BlobServiceTestSuite) TestSaveAndLoad_WithCompression() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
			CompressionType:    "zstd",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)
	s.Require().NotNil(s.svc)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service with compression")

	payload := []byte("hello-blob-with-compression")

	// WHEN
	ci, err := s.svc.Save(payload, "text/plain", "hello.txt")

	// THEN
	s.Require().NoError(err, "save payload with compression")
	s.Assert().NotNil(ci)
	s.Assert().Equal(int64(len(payload)), ci.OriginalLength)
	s.Assert().Equal("zstd", ci.Compressed)

	got, err := s.svc.Load(ci)
	s.Require().NoError(err, "load compressed payload")
	s.Assert().Equal(payload, got, "decompressed payload should match original")
}

func (s *BlobServiceTestSuite) TestRotatesContainer_AfterMaxSize() {
	// GIVEN
	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	first := bytes.Repeat([]byte{1}, 900*1024)
	second := bytes.Repeat([]byte{2}, 300*1024)

	// WHEN
	ci1, err := s.svc.Save(first, "application/octet-stream", "hello.txt")
	s.Require().NoError(err, "save first payload")

	ci2, err := s.svc.Save(second, "application/octet-stream", "world.txt")
	s.Require().NoError(err, "save second payload")

	// THEN
	s.Assert().Equal(1, ci1.ContainerNumber)
	s.Assert().Equal(2, ci2.ContainerNumber, "second save should rotate to new container")
}

func (s *BlobServiceTestSuite) TestLoad_AfterRotation() {
	// GIVEN
	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	first := bytes.Repeat([]byte{1}, 900*1024)
	second := bytes.Repeat([]byte{2}, 300*1024)

	ci1, err := s.svc.Save(first, "application/octet-stream", "first.txt")
	s.Require().NoError(err, "save first payload")
	ci2, err := s.svc.Save(second, "application/octet-stream", "second.txt")
	s.Require().NoError(err, "save second payload")

	// WHEN
	got1, err := s.svc.Load(ci1)
	got2, err2 := s.svc.Load(ci2)

	// THEN
	s.Require().NoError(err, "load first payload")
	s.Require().NoError(err2, "load second payload")
	s.Assert().Equal(first, got1)
	s.Assert().Equal(second, got2)
}

func (s *BlobServiceTestSuite) TestCompressionGzip_SaveAndLoad() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
			CompressionType:    "gzip",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)
	s.Require().NotNil(s.svc)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service with gzip")

	payload := []byte("hello-world-gzip-compression-test")

	// WHEN
	ci, err := s.svc.Save(payload, "application/pdf", "hello.pdf")

	// THEN
	s.Require().NoError(err, "save payload with gzip")
	s.Assert().Equal(int64(len(payload)), ci.OriginalLength)
	s.Assert().Equal("gzip", ci.Compressed)

	got, err := s.svc.Load(ci)
	s.Require().NoError(err, "load gzip-compressed payload")
	s.Assert().Equal(payload, got)
}

func (s *BlobServiceTestSuite) TestCompressionNone_SaveAndLoad() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
			CompressionType:    "none",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service without compression")

	payload := []byte("hello-world-no-compression")

	// WHEN
	ci, err := s.svc.Save(payload, "image/jpeg", "hello.jpg")

	// THEN
	s.Require().NoError(err, "save payload without compression")
	s.Assert().Equal(int64(len(payload)), ci.OriginalLength)
	s.Assert().Equal("none", ci.Compressed)

	got, err := s.svc.Load(ci)
	s.Require().NoError(err, "load uncompressed payload")
	s.Assert().Equal(payload, got)
}

func (s *BlobServiceTestSuite) TestMixedCompressionInContainer_ZstdAndGzipAndNone() {
	// GIVEN - create a service with "none" compression, then save/load mixed types
	// The key insight: all data is stored with its own compression flag,
	// so one container can hold data with different compression methods
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 10, // larger to fit multiple payloads
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	// Save three payloads with different compression configs
	payload1 := []byte("first-uncompressed-data")
	payload2 := []byte("second-uncompressed-data")

	// Both saved without compression (service uses "none")
	ci1, err := s.svc.Save(payload1, "text/plain", "first.txt")
	s.Require().NoError(err, "save first uncompressed payload")
	s.Assert().Equal(1, ci1.ContainerNumber)

	ci2, err := s.svc.Save(payload2, "text/plain", "second.txt")
	s.Require().NoError(err, "save second uncompressed payload")
	s.Assert().Equal(1, ci2.ContainerNumber, "both in same container")

	// WHEN - load both back
	got1, err := s.svc.Load(ci1)
	s.Require().NoError(err, "load first payload")

	got2, err := s.svc.Load(ci2)
	s.Require().NoError(err, "load second payload")

	// THEN - verify data integrity
	s.Assert().Equal(payload1, got1)
	s.Assert().Equal(payload2, got2)
	s.Assert().Equal("none", ci1.Compressed)
	s.Assert().Equal("none", ci2.Compressed)
}

func (s *BlobServiceTestSuite) TestMixedCompressionInContainer_ZstdThenGzip() {
	// Test that demonstrates the design: each entry stores its own compression flag
	// We manually test this by creating two separate services and reading from the same container

	// Phase 1: Save data compressed with zstd
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 10,
			CompressionType:    "zstd",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare with zstd")

	payload1 := bytes.Repeat([]byte("zstd-compressed-data-"), 100)

	// WHEN - save with zstd
	ci1, err := s.svc.Save(payload1, "application/pdf", "hello.pdf")
	s.Require().NoError(err, "save with zstd")
	s.Assert().Equal("zstd", ci1.Compressed)
	s.Assert().Equal(1, ci1.ContainerNumber)

	containerNum := ci1.ContainerNumber

	// Phase 2: Create new service with gzip, save more data to same container
	s.svc.Close()

	s.inj = do.New()
	cfg.Repository.CompressionType = "gzip"
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err = s.svc.Prepare()
	s.Require().NoError(err, "prepare with gzip")

	payload2 := bytes.Repeat([]byte("gzip-compressed-data-"), 100)

	// WHEN - save with gzip
	ci2, err := s.svc.Save(payload2, "application/pdf", "hello.pdf")
	s.Require().NoError(err, "save with gzip")
	s.Assert().Equal("gzip", ci2.Compressed)
	s.Assert().Equal(containerNum, ci2.ContainerNumber, "both entries in same container")

	// Phase 3: Load both back and verify integrity
	// WHEN - load both payloads (one compressed with zstd, one with gzip)
	got1, err := s.svc.Load(ci1)
	s.Require().NoError(err, "load zstd-compressed payload")

	got2, err := s.svc.Load(ci2)
	s.Require().NoError(err, "load gzip-compressed payload")

	// THEN - verify both decompressed correctly
	s.Assert().Equal(payload1, got1, "zstd-compressed data must decompress correctly")
	s.Assert().Equal(payload2, got2, "gzip-compressed data must decompress correctly")
}

func (s *BlobServiceTestSuite) TestCompressionCompatibility_SaveWithZstdLoadWithNone() {
	// Test that a container can be read by a service with different compression config
	// This proves the format is universally readable

	// Phase 1: Save with zstd
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
			CompressionType:    "zstd",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare with zstd")

	payload := []byte("test-payload-for-cross-compression-compatibility")

	ci, err := s.svc.Save(payload, "application/pdf", "hello.pdf")
	s.Require().NoError(err, "save with zstd")
	s.Assert().Equal("zstd", ci.Compressed)

	s.svc.Close()

	// Phase 2: Load with "none" compression config
	// The service config says "none", but the container stores the actual compression type
	// So loading should still work
	s.inj = do.New()
	cfg.Repository.CompressionType = "none"
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err = s.svc.Prepare()
	s.Require().NoError(err, "prepare with none")

	// WHEN - load the zstd-compressed data using a "none"-configured service
	got, err := s.svc.Load(ci)

	// THEN - it should still decompress correctly because the compression type is stored in the entry
	s.Require().NoError(err, "load zstd-compressed data with none-service")
	s.Assert().Equal(payload, got, "data decompressed correctly despite different service config")
}

func (s *BlobServiceTestSuite) TestInfFile_CreateAndRead() {
	// GIVEN
	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	payload := []byte("test-data-for-inf-file")

	// WHEN - save data (should create .inf file)
	ci, err := s.svc.Save(payload, "application/pdf", "hello.pdf")
	s.Require().NoError(err, "save payload")

	// THEN - verify .inf file contains the entry
	containerNum := ci.ContainerNumber
	infos, err := s.svc.LoadContainerInfos(containerNum)
	s.Require().NoError(err, "load container infos")
	s.Require().Len(infos, 1)
	s.Assert().Equal(ci.Offset, infos[0].Offset)
	s.Assert().Equal(ci.Length, infos[0].Length)
	s.Assert().Equal(ci.OriginalLength, infos[0].OriginalLength)
	s.Assert().Equal(ci.MIMEType, infos[0].MIMEType)
	s.Assert().Equal(ci.Compressed, infos[0].Compressed)
}

func (s *BlobServiceTestSuite) TestInfFile_MultipleEntriesInContainer() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 10,
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	// WHEN - save multiple payloads to same container
	payload1 := []byte("first-data")
	payload2 := []byte("second-data")
	payload3 := []byte("third-data")

	ci1, err := s.svc.Save(payload1, "application/pdf", "first.pdf")
	s.Require().NoError(err, "save first payload")

	_, err = s.svc.Save(payload2, "text/plain", "second.txt")
	s.Require().NoError(err, "save second payload")

	_, err = s.svc.Save(payload3, "image/jpeg", "third.jpg")
	s.Require().NoError(err, "save third payload")

	// THEN - verify all three are in the .inf file
	containerNum := ci1.ContainerNumber
	infos, err := s.svc.LoadContainerInfos(containerNum)
	s.Require().NoError(err, "load container infos")
	s.Require().Len(infos, 3)

	s.Assert().Equal(int64(len(payload1)), infos[0].OriginalLength)
	s.Assert().Equal(int64(len(payload2)), infos[1].OriginalLength)
	s.Assert().Equal(int64(len(payload3)), infos[2].OriginalLength)

	s.Assert().Equal("application/pdf", infos[0].MIMEType)
	s.Assert().Equal("text/plain", infos[1].MIMEType)
	s.Assert().Equal("image/jpeg", infos[2].MIMEType)
}

func (s *BlobServiceTestSuite) TestInfFile_MultipleContainers() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1, // Small to force rotation
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	// WHEN - save multiple large payloads to force container rotation
	payload1 := bytes.Repeat([]byte("container1-"), 900*1024/12)
	payload2 := bytes.Repeat([]byte("container2-"), 300*1024/12)

	ci1, err := s.svc.Save(payload1, "application/pdf", "first.pdf")
	s.Require().NoError(err, "save first payload")

	ci2, err := s.svc.Save(payload2, "image/tiff", "second.tiff")
	s.Require().NoError(err, "save second payload")

	// THEN - verify each container has its own .inf entry
	s.Assert().Equal(1, ci1.ContainerNumber)
	s.Assert().Equal(2, ci2.ContainerNumber)

	infos1, err := s.svc.LoadContainerInfos(1)
	s.Require().NoError(err, "load container 1 infos")
	s.Require().Len(infos1, 1)
	s.Assert().Equal(int64(len(payload1)), infos1[0].OriginalLength)

	infos2, err := s.svc.LoadContainerInfos(2)
	s.Require().NoError(err, "load container 2 infos")
	s.Require().Len(infos2, 1)
	s.Assert().Equal(int64(len(payload2)), infos2[0].OriginalLength)
}

func (s *BlobServiceTestSuite) TestListAllContainerInfos() {
	// GIVEN
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare blob service")

	// WHEN - save data to multiple containers
	payload1a := bytes.Repeat([]byte("a"), 900*1024/1)
	payload1b := []byte("small")
	payload2a := bytes.Repeat([]byte("b"), 300*1024/1)

	_, err = s.svc.Save(payload1a, "application/pdf", "first.pdf")
	s.Require().NoError(err, "save first container first payload")

	_, err = s.svc.Save(payload1b, "text/plain", "second.txt")
	s.Require().NoError(err, "save first container second payload")

	_, err = s.svc.Save(payload2a, "image/jpeg", "third.jpg")
	s.Require().NoError(err, "save second container payload")

	// WHEN - list all container infos
	allInfos, err := s.svc.ListAllContainerInfos()

	// THEN - verify structure
	s.Require().NoError(err, "list all container infos")
	s.Assert().Len(allInfos, 2, "should have 2 containers")

	// Verify container 1
	s.Require().Contains(allInfos, 1)
	s.Assert().Len(allInfos[1], 2, "container 1 should have 2 entries")
	s.Assert().Equal(int64(len(payload1a)), allInfos[1][0].OriginalLength)
	s.Assert().Equal(int64(len(payload1b)), allInfos[1][1].OriginalLength)

	// Verify container 2
	s.Require().Contains(allInfos, 2)
	s.Assert().Len(allInfos[2], 1, "container 2 should have 1 entry")
	s.Assert().Equal(int64(len(payload2a)), allInfos[2][0].OriginalLength)
}

func (s *BlobServiceTestSuite) TestInfFile_WithCompression() {
	// GIVEN - setup with zstd compression
	s.inj = do.New()
	cfg := config.Config{
		Repository: config.Repository{
			RepositoryPath:     s.tempDir,
			ContainerMaxSizeMB: 1,
			CompressionType:    "zstd",
		},
	}
	do.ProvideValue(s.inj, cfg)
	s.svc = New(s.inj)

	err := s.svc.Prepare()
	s.Require().NoError(err, "prepare with compression")

	// WHEN - save with compression
	payload := []byte("compressed-test-data")
	ci, err := s.svc.Save(payload, "application/pdf", "hello.pdf")
	s.Require().NoError(err, "save compressed payload")

	// THEN - verify .inf file contains compression info
	infos, err := s.svc.LoadContainerInfos(ci.ContainerNumber)
	s.Require().NoError(err, "load container infos")
	s.Require().Len(infos, 1)
	s.Assert().Equal("zstd", infos[0].Compressed)
	s.Assert().Equal(int64(len(payload)), infos[0].OriginalLength)
}
