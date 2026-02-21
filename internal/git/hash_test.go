package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashDirDeterministic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("world"), 0644)

	h1, err := HashDir(dir)
	if err != nil {
		t.Fatalf("first hash: %v", err)
	}
	h2, err := HashDir(dir)
	if err != nil {
		t.Fatalf("second hash: %v", err)
	}
	if h1 != h2 {
		t.Errorf("not deterministic: %s != %s", h1, h2)
	}
	if !strings.HasPrefix(h1, "sha256:") {
		t.Errorf("missing sha256 prefix: %s", h1)
	}
}

func TestHashDirDifferentContent(t *testing.T) {
	d1 := t.TempDir()
	os.WriteFile(filepath.Join(d1, "f.txt"), []byte("aaa"), 0644)

	d2 := t.TempDir()
	os.WriteFile(filepath.Join(d2, "f.txt"), []byte("bbb"), 0644)

	h1, _ := HashDir(d1)
	h2, _ := HashDir(d2)
	if h1 == h2 {
		t.Error("different content produced same hash")
	}
}

func TestHashDirDifferentFilenames(t *testing.T) {
	d1 := t.TempDir()
	os.WriteFile(filepath.Join(d1, "a.txt"), []byte("same"), 0644)

	d2 := t.TempDir()
	os.WriteFile(filepath.Join(d2, "b.txt"), []byte("same"), 0644)

	h1, _ := HashDir(d1)
	h2, _ := HashDir(d2)
	if h1 == h2 {
		t.Error("different filenames produced same hash")
	}
}

func TestHashDirSubdirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("deep"), 0644)

	h, err := HashDir(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if h == "" {
		t.Error("empty hash")
	}
}

func TestHashDirEmpty(t *testing.T) {
	dir := t.TempDir()
	h, err := HashDir(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasPrefix(h, "sha256:") {
		t.Errorf("empty dir hash: %s", h)
	}
}

func TestHashDirNotFound(t *testing.T) {
	_, err := HashDir("/nonexistent/dir")
	if err == nil {
		t.Error("expected error")
	}
}
