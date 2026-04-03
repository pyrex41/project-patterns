package add

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/pyrex41/project-patterns/internal/config"
	gh "github.com/pyrex41/project-patterns/internal/github"
	"github.com/pyrex41/project-patterns/internal/project"
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
	return runRepoWith(cmd, args[0])
}

func runRepoWith(cmd *cobra.Command, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL %q: must be an absolute URL (e.g. https://github.com/user/repo)", rawURL)
	}

	urlPath := parsed.Path
	repoName := path.Base(urlPath)
	repoName = strings.TrimSuffix(repoName, ".git")

	slug := ownerRepoSlug(urlPath)

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = repoName
	}

	tags, _ := cmd.Flags().GetStringSlice("tags")
	desc, _ := cmd.Flags().GetString("desc")

	// Fetch description from README if not provided (GitHub only).
	if desc == "" && strings.Contains(parsed.Host, "github.com") {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 2 {
			owner := parts[0]
			repo := strings.TrimSuffix(parts[1], ".git")
			cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, loadErr := config.Load(cfgPath)
			token := ""
			if loadErr == nil {
				token = cfg.GitHubToken
			}
			client := gh.NewClient(token)
			desc = client.FetchReadmeFirstParagraph(owner, repo)
		}
	}

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
func ownerRepoSlug(urlPath string) string {
	p := strings.Trim(urlPath, "/")
	p = strings.TrimSuffix(p, ".git")
	return strings.ReplaceAll(p, "/", "-")
}
