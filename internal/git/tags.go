package git

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Tag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

// ListTags returns all tags for a repo via gh api.
func ListTags(owner, repo string) ([]Tag, error) {
	out, err := RunGH("api", fmt.Sprintf("repos/%s/%s/tags", owner, repo), "--paginate")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	if err := json.Unmarshal([]byte(out), &tags); err != nil {
		return nil, fmt.Errorf("parsing tags: %w", err)
	}
	return tags, nil
}

// ResolveVersion finds the best tag matching a semver range.
// Supports ^, ~, exact, and * ranges.
func ResolveVersion(tags []Tag, versionRange string) (*Tag, error) {
	var semverTags []struct {
		Tag   Tag
		Major int
		Minor int
		Patch int
	}

	for _, t := range tags {
		name := strings.TrimPrefix(t.Name, "v")
		var major, minor, patch int
		n, _ := fmt.Sscanf(name, "%d.%d.%d", &major, &minor, &patch)
		if n == 3 {
			semverTags = append(semverTags, struct {
				Tag   Tag
				Major int
				Minor int
				Patch int
			}{t, major, minor, patch})
		}
	}

	if len(semverTags) == 0 {
		return nil, fmt.Errorf("no semver tags found")
	}

	// Sort descending
	sort.Slice(semverTags, func(i, j int) bool {
		if semverTags[i].Major != semverTags[j].Major {
			return semverTags[i].Major > semverTags[j].Major
		}
		if semverTags[i].Minor != semverTags[j].Minor {
			return semverTags[i].Minor > semverTags[j].Minor
		}
		return semverTags[i].Patch > semverTags[j].Patch
	})

	if versionRange == "*" || versionRange == "latest" {
		return &semverTags[0].Tag, nil
	}

	var minMajor, minMinor, minPatch int
	var maxMajor, maxMinor, maxPatch int
	maxMajor = 999
	maxMinor = 999
	maxPatch = 999

	clean := versionRange
	mode := "exact"
	if strings.HasPrefix(clean, "^") {
		mode = "caret"
		clean = strings.TrimPrefix(clean, "^")
	} else if strings.HasPrefix(clean, "~") {
		mode = "tilde"
		clean = strings.TrimPrefix(clean, "~")
	} else if strings.HasPrefix(clean, ">=") {
		mode = "gte"
		clean = strings.TrimPrefix(clean, ">=")
	}

	fmt.Sscanf(clean, "%d.%d.%d", &minMajor, &minMinor, &minPatch)

	switch mode {
	case "caret":
		maxMajor = minMajor
	case "tilde":
		maxMajor = minMajor
		maxMinor = minMinor
	case "exact":
		maxMajor = minMajor
		maxMinor = minMinor
		maxPatch = minPatch
	case "gte":
		// no upper bound
	}

	for _, st := range semverTags {
		if st.Major < minMajor || (st.Major == minMajor && st.Minor < minMinor) || (st.Major == minMajor && st.Minor == minMinor && st.Patch < minPatch) {
			continue
		}
		if mode == "caret" && st.Major > maxMajor {
			continue
		}
		if mode == "tilde" && (st.Major > maxMajor || (st.Major == maxMajor && st.Minor > maxMinor)) {
			continue
		}
		if mode == "exact" && (st.Major != maxMajor || st.Minor != maxMinor || st.Patch != maxPatch) {
			continue
		}
		return &st.Tag, nil
	}

	return nil, fmt.Errorf("no tag matching %q", versionRange)
}
