package add

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenb/project-patterns/internal/config"
	"github.com/reubenb/project-patterns/internal/project"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo <git-url>",
	Short: "Add a remote git repository as a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepo,
}

func init() {
	repoCmd.Flags().String("name", "", "project name (defaults to repo name from URL)")
	repoCmd.Flags().StringSlice("tags", nil, "tags to apply to the project")
	repoCmd.Flags().String("desc", "", "short description of the project")
	repoCmd.Flags().Bool("private", false, "mark project as private")
	AddCmd.AddCommand(repoCmd)
}

func runRepo(cmd *cobra.Command, args []string) error {
	rawURL := args[0]

	// Parse and validate the URL.
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL %q: must be an absolute URL (e.g. https://github.com/user/repo)", rawURL)
	}

	// Derive the repo name from the URL path (last segment, strip .git).
	urlPath := parsed.Path
	repoName := path.Base(urlPath)
	repoName = strings.TrimSuffix(repoName, ".git")

	// Derive owner/repo slug for the clone path (last two path segments).
	slug := ownerRepoSlug(urlPath)

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = repoName
	}

	tags, _ := cmd.Flags().GetStringSlice("tags")
	desc, _ := cmd.Flags().GetString("desc")

	id := project.GenerateID(name)

	cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	clonePath := config.ExpandPath(cfg.CacheDir + "/" + slug)

	p := project.Project{
		ID:          id,
		Name:        name,
		Type:        "git",
		URL:         rawURL,
		ClonePath:   clonePath,
		Tags:        tags,
		Description: desc,
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	updated := cfg.AddOrUpdateProject(p)

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if updated {
		color.Yellow("Updated project: %s (%s)", name, id)
	} else {
		color.Green("Added project: %s (%s)", name, id)
	}

	return nil
}

// ownerRepoSlug returns an "owner-repo" style slug derived from the URL path.
// For "/user/repo" or "/user/repo.git" it returns "user-repo".
func ownerRepoSlug(urlPath string) string {
	// Strip leading/trailing slashes and .git suffix.
	p := strings.Trim(urlPath, "/")
	p = strings.TrimSuffix(p, ".git")
	// Replace path separators with hyphens.
	return strings.ReplaceAll(p, "/", "-")
}
