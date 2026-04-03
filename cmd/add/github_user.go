package add

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/pyrex41/project-patterns/internal/config"
	gh "github.com/pyrex41/project-patterns/internal/github"
	"github.com/pyrex41/project-patterns/internal/project"
	"github.com/spf13/cobra"
)

var githubUserCmd = &cobra.Command{
	Use:   "github-user <username>",
	Short: "Index public repos for a GitHub user",
	Args:  cobra.ExactArgs(1),
	RunE:  runGitHubUser,
}

func init() {
	githubUserCmd.Flags().Bool("include-private", false, "include private repositories (requires a GitHub token)")
	githubUserCmd.Flags().StringSlice("tags", nil, "default tags to apply to all imported projects")
	AddCmd.AddCommand(githubUserCmd)
}

func runGitHubUser(cmd *cobra.Command, args []string) error {
	return runGitHubUserWith(cmd, args[0])
}

func runGitHubUserWith(cmd *cobra.Command, input string) error {
	// input may be a full URL like https://github.com/user — extract the username.
	username := input
	if strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "http://") {
		u, err := url.Parse(input)
		if err == nil {
			username = strings.Trim(u.Path, "/")
		}
	}

	includePrivate, _ := cmd.Flags().GetBool("include-private")
	defaultTags, _ := cmd.Flags().GetStringSlice("tags")

	cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client := gh.NewClient(cfg.GitHubToken)

	repos, err := client.ListUserRepos(username, includePrivate)
	if err != nil {
		return fmt.Errorf("listing repos for %s: %w", username, err)
	}

	added := 0
	updated := 0

	for _, repo := range repos {
		if repo.Fork || repo.Archived {
			continue
		}

		tags := make([]string, 0, len(repo.Topics)+len(defaultTags))
		tags = append(tags, repo.Topics...)
		tags = append(tags, defaultTags...)
		tags = uniqueStrings(tags)

		id := project.GenerateID(repo.Name)
		slug := strings.ReplaceAll(repo.FullName, "/", "-")
		clonePath := config.ExpandPath(cfg.CacheDir + "/" + slug)

		p := project.Project{
			ID:          id,
			Name:        repo.Name,
			Type:        "git",
			URL:         repo.HTMLURL,
			ClonePath:   clonePath,
			Description: repo.Description,
			Tags:        tags,
		}

		if err := p.Validate(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "skipping %s: %v\n", repo.FullName, err)
			continue
		}

		wasUpdated := cfg.AddOrUpdateProject(p)
		if wasUpdated {
			updated++
		} else {
			added++
		}
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	color.Green("Added %d projects, updated %d projects from %s", added, updated, username)
	return nil
}

// uniqueStrings returns a deduplicated slice preserving order.
func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
