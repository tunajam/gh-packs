package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.yaml")

	yaml := `name: "@org/svc"
version: "1.0.0"
description: "test service"
pack:
  type: service
dependencies:
  "@org/payments":
    version: "^1.0.0"
    source: "github.com/org/payments"
    reason: "billing"
resolve:
  max_depth: 3
`
	os.WriteFile(path, []byte(yaml), 0644)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Name != "@org/svc" {
		t.Errorf("Name = %q, want @org/svc", m.Name)
	}
	if m.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", m.Version)
	}
	if m.Resolve.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", m.Resolve.MaxDepth)
	}
	dep, ok := m.Dependencies["@org/payments"]
	if !ok {
		t.Fatal("missing @org/payments dependency")
	}
	if dep.Version != "^1.0.0" {
		t.Errorf("dep version = %q", dep.Version)
	}
	if dep.Reason != "billing" {
		t.Errorf("dep reason = %q", dep.Reason)
	}
}

func TestLoadManifestDefaultMaxDepth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.yaml")
	os.WriteFile(path, []byte("name: test\nversion: \"1.0.0\"\n"), 0644)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Resolve.MaxDepth != 2 {
		t.Errorf("default MaxDepth = %d, want 2", m.Resolve.MaxDepth)
	}
}

func TestLoadManifestNotFound(t *testing.T) {
	_, err := Load("/nonexistent/packs.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadManifestInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.yaml")
	os.WriteFile(path, []byte("{{invalid"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "packs.yaml")

	m := &Manifest{
		Name:    "@test/svc",
		Version: "2.0.0",
		Dependencies: map[string]Dependency{
			"@org/api": {Version: "^1.0.0", Source: "github.com/org/api"},
		},
	}
	if err := m.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Name != "@test/svc" {
		t.Errorf("Name = %q", loaded.Name)
	}
	if _, ok := loaded.Dependencies["@org/api"]; !ok {
		t.Error("missing dependency after round-trip")
	}
}

func TestContextPathDefault(t *testing.T) {
	d := Dependency{}
	if d.ContextPath() != ".context" {
		t.Errorf("default ContextPath = %q, want .context", d.ContextPath())
	}
}

func TestContextPathCustom(t *testing.T) {
	d := Dependency{Path: "docs/context"}
	if d.ContextPath() != "docs/context" {
		t.Errorf("ContextPath = %q, want docs/context", d.ContextPath())
	}
}
