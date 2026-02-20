package git

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
}

type TreeResponse struct {
	SHA  string      `json:"sha"`
	Tree []TreeEntry `json:"tree"`
}

type ContentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	SHA      string `json:"sha"`
}

// FetchContextDir downloads the context directory from a repo at a specific ref.
// Returns the list of files written.
func FetchContextDir(owner, repo, ref, contextPath, destDir string) ([]string, error) {
	// Get tree at ref
	out, err := RunGH("api", fmt.Sprintf("repos/%s/%s/git/trees/%s?recursive=1", owner, repo, ref))
	if err != nil {
		return nil, fmt.Errorf("fetching tree: %w", err)
	}

	var tree TreeResponse
	if err := json.Unmarshal([]byte(out), &tree); err != nil {
		return nil, fmt.Errorf("parsing tree: %w", err)
	}

	// Filter to context path
	prefix := contextPath + "/"
	var files []string
	for _, entry := range tree.Tree {
		if entry.Type == "blob" && (entry.Path == contextPath || strings.HasPrefix(entry.Path, prefix)) {
			files = append(files, entry.Path)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found at %s/ in %s/%s@%s", contextPath, owner, repo, ref)
	}

	// Download each file
	var written []string
	for _, f := range files {
		out, err := RunGH("api", fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, f, ref))
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", f, err)
		}

		var content ContentResponse
		if err := json.Unmarshal([]byte(out), &content); err != nil {
			return nil, fmt.Errorf("parsing content for %s: %w", f, err)
		}

		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", f, err)
		}

		// Strip context path prefix, write under destDir
		relPath := strings.TrimPrefix(f, prefix)
		if relPath == contextPath {
			relPath = filepath.Base(f)
		}
		destPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(destPath, decoded, 0644); err != nil {
			return nil, err
		}
		written = append(written, destPath)
	}

	return written, nil
}

// FetchFile downloads a single file from a repo at a ref.
func FetchFile(owner, repo, ref, path string) ([]byte, error) {
	out, err := RunGH("api", fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, path, ref))
	if err != nil {
		return nil, err
	}
	var content ContentResponse
	if err := json.Unmarshal([]byte(out), &content); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
}
