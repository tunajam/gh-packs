package agent

import (
	"strings"
	"testing"
)

func TestDefaultResolvePromptNotEmpty(t *testing.T) {
	if DefaultResolvePrompt == "" {
		t.Error("DefaultResolvePrompt is empty")
	}
}

func TestDefaultResolvePromptContent(t *testing.T) {
	required := []string{"context resolution", "dependency manifest", "briefing"}
	for _, s := range required {
		if !strings.Contains(strings.ToLower(DefaultResolvePrompt), s) {
			t.Errorf("prompt missing expected content: %q", s)
		}
	}
}
