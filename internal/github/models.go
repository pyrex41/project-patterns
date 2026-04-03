package github

import "time"

// Repository represents a GitHub repository from the REST API v3
// /users/{user}/repos response.
type Repository struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Private       bool     `json:"private"`
	HTMLURL       string   `json:"html_url"`
	CloneURL      string   `json:"clone_url"`
	Description   string   `json:"description"`
	Fork          bool     `json:"fork"`
	Archived      bool     `json:"archived"`
	Language      string   `json:"language"`
	Topics        []string `json:"topics"`
	DefaultBranch string   `json:"default_branch"`
}

// String returns the full name of the repository (e.g. "owner/repo").
func (r Repository) String() string {
	return r.FullName
}

// RateLimitInfo holds GitHub API rate limit state parsed from response headers.
type RateLimitInfo struct {
	Remaining int
	Limit     int
	ResetAt   time.Time
}
