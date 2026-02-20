package manifest

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const LockFile = "packs.lock"

type Lockfile struct {
	LockfileVersion int                    `yaml:"lockfile_version"`
	ResolvedAt      string                 `yaml:"resolved_at"`
	Packages        map[string]LockedPack  `yaml:"packages"`
}

type LockedPack struct {
	Source      string `yaml:"source"`
	Ref         string `yaml:"ref"`
	Commit      string `yaml:"commit"`
	ContentHash string `yaml:"content_hash"`
}

func LoadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading lockfile: %w", err)
	}
	var l Lockfile
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("parsing lockfile: %w", err)
	}
	return &l, nil
}

func (l *Lockfile) Save(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshaling lockfile: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func NewLockfile() *Lockfile {
	return &Lockfile{
		LockfileVersion: 1,
		ResolvedAt:      time.Now().UTC().Format(time.RFC3339),
		Packages:        make(map[string]LockedPack),
	}
}
