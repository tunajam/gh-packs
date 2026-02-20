package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const ManifestFile = "packs.yaml"

type Manifest struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Description  string                `yaml:"description,omitempty"`
	Pack         PackInfo              `yaml:"pack,omitempty"`
	Dependencies map[string]Dependency `yaml:"dependencies,omitempty"`
	Resolve      ResolveConfig         `yaml:"resolve,omitempty"`
}

type PackInfo struct {
	Type  string `yaml:"type,omitempty"`
	Proto string `yaml:"proto,omitempty"`
}

type Dependency struct {
	Version   string   `yaml:"version"`
	Source    string   `yaml:"source"`
	Path      string   `yaml:"path,omitempty"`
	Reason    string   `yaml:"reason,omitempty"`
	Relevance []string `yaml:"relevance,omitempty"`
}

func (d Dependency) ContextPath() string {
	if d.Path != "" {
		return d.Path
	}
	return ".context"
}

type ResolveConfig struct {
	MaxDepth           int    `yaml:"max_depth,omitempty"`
	SystemPromptAppend string `yaml:"system_prompt_append,omitempty"`
}

func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if m.Resolve.MaxDepth == 0 {
		m.Resolve.MaxDepth = 2
	}
	return &m, nil
}

func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
