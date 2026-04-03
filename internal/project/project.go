package project

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Project represents a tracked project entry.
type Project struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`                   // "local" or "git"
	Path        string   `yaml:"path,omitempty"`         // for local type
	URL         string   `yaml:"url,omitempty"`          // for git type
	ClonePath   string   `yaml:"clone_path,omitempty"`   // managed local clone
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description,omitempty"`
	LastSynced  string   `yaml:"last_synced,omitempty"` // RFC3339 timestamp
}

// Validate checks that the project has all required fields set correctly.
func (p *Project) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("project id is required")
	}
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}
	switch p.Type {
	case "local":
		if p.Path == "" {
			return fmt.Errorf("project of type %q must have a path", p.Type)
		}
	case "git":
		if p.URL == "" {
			return fmt.Errorf("project of type %q must have a url", p.Type)
		}
	default:
		return fmt.Errorf("project type must be %q or %q, got %q", "local", "git", p.Type)
	}
	return nil
}

var (
	nonAlphanumRe  = regexp.MustCompile(`[^a-z0-9]+`)
	multiHyphenRe  = regexp.MustCompile(`-{2,}`)
)

// GenerateID converts a project name into a URL-friendly slug.
// It lowercases the name, replaces spaces and special characters with hyphens,
// collapses consecutive hyphens, and trims hyphens from the edges.
func GenerateID(name string) string {
	// Normalise to ASCII-friendly lowercase by mapping non-letter/digit runes to hyphens.
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	slug := b.String()
	slug = nonAlphanumRe.ReplaceAllString(slug, "-")
	slug = multiHyphenRe.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// HasTag reports whether the project is tagged with the given tag (case-insensitive).
func (p *Project) HasTag(tag string) bool {
	needle := strings.ToLower(tag)
	for _, t := range p.Tags {
		if strings.ToLower(t) == needle {
			return true
		}
	}
	return false
}

// HasAllTags reports whether the project has every one of the given tags (AND logic, case-insensitive).
func (p *Project) HasAllTags(tags []string) bool {
	for _, tag := range tags {
		if !p.HasTag(tag) {
			return false
		}
	}
	return true
}

// MatchesQuery reports whether the query string appears (case-insensitively) in the
// project Name or Description.
func (p *Project) MatchesQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(p.Name), q) ||
		strings.Contains(strings.ToLower(p.Description), q)
}

// LocalPath returns the resolved local directory for this project.
// It returns Path if set, otherwise ClonePath.
func (p *Project) LocalPath() string {
	if p.Path != "" {
		return p.Path
	}
	return p.ClonePath
}
