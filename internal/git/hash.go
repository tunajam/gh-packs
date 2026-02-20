package git

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// HashDir computes a SHA-256 hash of all file contents in a directory.
func HashDir(dir string) (string, error) {
	h := sha256.New()
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(paths)
	for _, p := range paths {
		data, err := os.ReadFile(filepath.Join(dir, p))
		if err != nil {
			return "", err
		}
		h.Write([]byte(p))
		h.Write(data)
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}
