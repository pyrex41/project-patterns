package git

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildAuthURL inserts token as userinfo into rawURL so that git can
// authenticate against GitHub without a credential helper.  If token is
// empty the original URL is returned unchanged.  The returned URL is
// guaranteed to end with ".git".
func BuildAuthURL(rawURL, token string) string {
	if token == "" {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.User = url.User(token)

	result := u.String()
	if !strings.HasSuffix(result, ".git") {
		result += ".git"
	}

	return result
}

// IsGitRepo reports whether path contains a .git directory and therefore
// looks like an initialised git repository.
func IsGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

// Clone clones url into dest.  If token is non-empty it is embedded in the
// remote URL as userinfo so that private repositories can be fetched without
// interactive credential prompts.  The parent directory of dest is created
// when it does not already exist.
func Clone(rawURL, dest, token string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", dest, err)
	}

	cloneURL := rawURL
	if token != "" {
		cloneURL = BuildAuthURL(rawURL, token)
	}

	cmd := exec.Command(gitPath, "clone", cloneURL, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %s: %w\n%s", rawURL, err, strings.TrimSpace(string(out)))
	}

	return nil
}

// Pull runs `git pull` inside repoPath, fetching and merging the upstream
// changes for the currently checked-out branch.
func Pull(repoPath string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	cmd := exec.Command(gitPath, "-C", repoPath, "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull in %s: %w\n%s", repoPath, err, strings.TrimSpace(string(out)))
	}

	return nil
}
