package git

import (
	"testing"
)

func makeTags(names ...string) []Tag {
	tags := make([]Tag, len(names))
	for i, n := range names {
		tags[i] = Tag{Name: n}
		tags[i].Commit.SHA = "sha-" + n
	}
	return tags
}

func TestResolveVersionExact(t *testing.T) {
	tags := makeTags("v1.0.0", "v1.1.0", "v2.0.0")
	tag, err := ResolveVersion(tags, "1.1.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v1.1.0" {
		t.Errorf("got %s, want v1.1.0", tag.Name)
	}
}

func TestResolveVersionExactNotFound(t *testing.T) {
	tags := makeTags("v1.0.0", "v2.0.0")
	_, err := ResolveVersion(tags, "1.5.0")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveVersionCaret(t *testing.T) {
	tags := makeTags("v1.0.0", "v1.2.0", "v1.9.5", "v2.0.0")
	tag, err := ResolveVersion(tags, "^1.0.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Should pick highest 1.x.x
	if tag.Name != "v1.9.5" {
		t.Errorf("got %s, want v1.9.5", tag.Name)
	}
}

func TestResolveVersionCaretExcludesNextMajor(t *testing.T) {
	tags := makeTags("v1.0.0", "v2.0.0", "v3.0.0")
	tag, err := ResolveVersion(tags, "^1.0.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v1.0.0" {
		t.Errorf("got %s, want v1.0.0", tag.Name)
	}
}

func TestResolveVersionTilde(t *testing.T) {
	tags := makeTags("v1.0.0", "v1.0.5", "v1.1.0", "v1.2.0")
	tag, err := ResolveVersion(tags, "~1.0.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Should pick highest 1.0.x
	if tag.Name != "v1.0.5" {
		t.Errorf("got %s, want v1.0.5", tag.Name)
	}
}

func TestResolveVersionTildeExcludesNextMinor(t *testing.T) {
	tags := makeTags("v1.0.0", "v1.1.0")
	tag, err := ResolveVersion(tags, "~1.0.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v1.0.0" {
		t.Errorf("got %s, want v1.0.0", tag.Name)
	}
}

func TestResolveVersionGte(t *testing.T) {
	tags := makeTags("v0.9.0", "v1.0.0", "v2.0.0", "v3.0.0")
	tag, err := ResolveVersion(tags, ">=1.0.0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v3.0.0" {
		t.Errorf("got %s, want v3.0.0", tag.Name)
	}
}

func TestResolveVersionGteExcludesLower(t *testing.T) {
	tags := makeTags("v0.1.0", "v0.9.0")
	_, err := ResolveVersion(tags, ">=1.0.0")
	if err == nil {
		t.Fatal("expected error, all tags below minimum")
	}
}

func TestResolveVersionWildcard(t *testing.T) {
	tags := makeTags("v1.0.0", "v2.0.0", "v3.0.0")
	tag, err := ResolveVersion(tags, "*")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v3.0.0" {
		t.Errorf("got %s, want v3.0.0", tag.Name)
	}
}

func TestResolveVersionLatest(t *testing.T) {
	tags := makeTags("v1.0.0", "v5.0.0")
	tag, err := ResolveVersion(tags, "latest")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tag.Name != "v5.0.0" {
		t.Errorf("got %s, want v5.0.0", tag.Name)
	}
}

func TestResolveVersionNoSemverTags(t *testing.T) {
	tags := makeTags("release-1", "nightly")
	_, err := ResolveVersion(tags, "^1.0.0")
	if err == nil {
		t.Fatal("expected error for non-semver tags")
	}
}

func TestResolveVersionEmptyTags(t *testing.T) {
	_, err := ResolveVersion(nil, "^1.0.0")
	if err == nil {
		t.Fatal("expected error for empty tags")
	}
}
