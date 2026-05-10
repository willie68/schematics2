package blob

import (
	"bytes"
	"testing"

	"github.com/willie68/schematic2/backend/internal/config"
)

func TestServiceSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	svc := New()
	t.Cleanup(func() {
		_ = svc.Close()
	})

	err := svc.Prepare(config.Repository{RepositoryPath: dir, ContainerMaxSizeMB: 1})
	if err != nil {
		t.Fatalf("prepare blob service: %v", err)
	}

	payload := []byte("hello-blob")
	ci, err := svc.Save(payload, "text/plain")
	if err != nil {
		t.Fatalf("save payload: %v", err)
	}

	if ci.ContainerNumber != 1 {
		t.Fatalf("expected container 1, got %d", ci.ContainerNumber)
	}
	if ci.Length != int64(len(payload)) {
		t.Fatalf("expected length %d, got %d", len(payload), ci.Length)
	}
	if ci.MIMEType != "text/plain" {
		t.Fatalf("expected mimetype text/plain, got %q", ci.MIMEType)
	}

	got, err := svc.Load(ci)
	if err != nil {
		t.Fatalf("load payload: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("loaded payload mismatch: got %q want %q", string(got), string(payload))
	}
}

func TestServiceRotatesContainer(t *testing.T) {
	dir := t.TempDir()
	svc := New()
	t.Cleanup(func() {
		_ = svc.Close()
	})

	err := svc.Prepare(config.Repository{RepositoryPath: dir, ContainerMaxSizeMB: 1})
	if err != nil {
		t.Fatalf("prepare blob service: %v", err)
	}

	first := bytes.Repeat([]byte{1}, 900*1024)
	ci1, err := svc.Save(first, "application/octet-stream")
	if err != nil {
		t.Fatalf("save first payload: %v", err)
	}

	second := bytes.Repeat([]byte{2}, 300*1024)
	ci2, err := svc.Save(second, "application/octet-stream")
	if err != nil {
		t.Fatalf("save second payload: %v", err)
	}

	if ci1.ContainerNumber != 1 {
		t.Fatalf("expected first container 1, got %d", ci1.ContainerNumber)
	}
	if ci2.ContainerNumber != 2 {
		t.Fatalf("expected second container 2 after rotation, got %d", ci2.ContainerNumber)
	}
}
