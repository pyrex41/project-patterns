package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a minimal GitHub REST API v3 client.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient returns a Client with a 30-second HTTP timeout.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListUserRepos fetches all repositories for the given username, following
// pagination automatically. When includePrivate is true and a token is set the
// authenticated /user/repos endpoint is used instead, which includes private
// repos owned by the authenticated user.
func (c *Client) ListUserRepos(username string, includePrivate bool) ([]Repository, error) {
	var all []Repository

	useAuthEndpoint := includePrivate && c.token != ""

	for page := 1; ; page++ {
		var rawURL string
		if useAuthEndpoint {
			rawURL = fmt.Sprintf(
				"https://api.github.com/user/repos?per_page=100&page=%d&affiliation=owner",
				page,
			)
		} else {
			rawURL = fmt.Sprintf(
				"https://api.github.com/users/%s/repos?per_page=100&page=%d",
				url.PathEscape(username),
				page,
			)
		}

		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("github: build request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("github: request failed (page %d): %w", page, err)
		}

		if err := c.checkRateLimit(resp); err != nil {
			resp.Body.Close()
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("github: unexpected status %d for page %d of %s", resp.StatusCode, page, rawURL)
		}

		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("github: decode response (page %d): %w", page, err)
		}
		resp.Body.Close()

		// Empty page signals end of results.
		if len(repos) == 0 {
			break
		}

		all = append(all, repos...)

		// Honour Link header if present; stop when there is no next page.
		if !hasNextPage(resp.Header.Get("Link")) {
			break
		}
	}

	return all, nil
}

// parseRateLimit extracts rate-limit state from the standard GitHub response
// headers.
func (c *Client) parseRateLimit(resp *http.Response) RateLimitInfo {
	remaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	limit, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Limit"))

	var resetAt time.Time
	if resetUnix, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		resetAt = time.Unix(resetUnix, 0)
	}

	return RateLimitInfo{
		Remaining: remaining,
		Limit:     limit,
		ResetAt:   resetAt,
	}
}

// checkRateLimit returns an error if the API rate limit has been exhausted.
func (c *Client) checkRateLimit(resp *http.Response) error {
	info := c.parseRateLimit(resp)
	if info.Remaining == 0 && info.Limit > 0 {
		return fmt.Errorf("github: rate limit exhausted; resets at %s", info.ResetAt.Format(time.RFC3339))
	}
	return nil
}

// BuildAuthCloneURL converts a plain HTTPS HTML URL such as
// https://github.com/user/repo into an authenticated clone URL of the form
// https://<token>@github.com/user/repo.git. When token is empty the original
// URL is returned (with .git appended if not already present).
func BuildAuthCloneURL(htmlURL string, token string) string {
	if !strings.HasSuffix(htmlURL, ".git") {
		htmlURL += ".git"
	}
	if token == "" {
		return htmlURL
	}

	// Insert token@ after the scheme (https://).
	const scheme = "https://"
	if strings.HasPrefix(htmlURL, scheme) {
		return scheme + token + "@" + htmlURL[len(scheme):]
	}

	return htmlURL
}

// hasNextPage reports whether a GitHub Link header contains a rel="next" entry.
func hasNextPage(link string) bool {
	if link == "" {
		return false
	}
	// Link header format: <url>; rel="next", <url>; rel="last"
	for _, part := range strings.Split(link, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			return true
		}
	}
	return false
}
