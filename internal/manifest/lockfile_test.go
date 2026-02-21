package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLockfile(t *testing.T) {
	l := NewLockfile()
	if l.LockfileVersion != 1 {
		t.Errorf("version = %d, want 1", l.LockfileVersion)
	}
	if l.ResolvedAt == "" {
		t.Error("ResolvedAt empty")
	}
	if l.Packages == nil {
		t.Error("Packages nil")
	}
}

func TestLockfileSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.lock")

	l := NewLockfile()
	l.Packages["@org/api@1.2.3"] = LockedPack{
		Source:      "github.com/org/api",
		Ref:         "v1.2.3",
		Commit:      "abc123",
		ContentHash: "sha256:deadbeef",
	}
	if err := l.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadLockfile(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.LockfileVersion != 1 {
		t.Errorf("version = %d", loaded.LockfileVersion)
	}
	pkg, ok := loaded.Packages["@org/api@1.2.3"]
	if !ok {
		t.Fatal("missing package")
	}
	if pkg.Commit != "abc123" {
		t.Errorf("commit = %q", pkg.Commit)
	}
	if pkg.ContentHash != "sha256:deadbeef" {
		t.Errorf("hash = %q", pkg.ContentHash)
	}
}

func TestLoadLockfileNotFound(t *testing.T) {
	_, err := LoadLockfile("/nonexistent/packs.lock")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadLockfileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.lock")
	os.WriteFile(path, []byte("{{bad"), 0644)

	_, err := LoadLockfile(path)
	if err == nil {
		t.Fatal("expected error")
	}
}
