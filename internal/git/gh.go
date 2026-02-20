package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunGH executes a gh CLI command and returns stdout.
func RunGH(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %s: %w", strings.Join(args, " "), stderr.String(), err)
	}
	return stdout.String(), nil
}

// ParseSource extracts owner/repo from a source like "github.com/org/repo".
func ParseSource(source string) (owner, repo string, err error) {
	source = strings.TrimPrefix(source, "https://")
	source = strings.TrimPrefix(source, "github.com/")
	parts := strings.SplitN(source, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid source %q: expected github.com/owner/repo", source)
	}
	return parts[0], parts[1], nil
}

// DetectRemote returns the owner/repo from the current git remote.
func DetectRemote() (owner, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("no git remote found: %w", err)
	}
	url := strings.TrimSpace(stdout.String())
	// Handle SSH: git@github.com:org/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		return ParseSource(url)
	}
	// Handle HTTPS
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, ".git")
	parts := strings.SplitN(url, "/", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
	}
	return parts[1], parts[2], nil
}
