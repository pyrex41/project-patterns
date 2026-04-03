package add

import (
	"fmt"
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
	username := args[0]

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
		// Skip forks and archived repos.
		if repo.Fork || repo.Archived {
			continue
		}

		// Build tags: repo topics + any default tags provided.
		tags := make([]string, 0, len(repo.Topics)+len(defaultTags))
		tags = append(tags, repo.Topics...)
		tags = append(tags, defaultTags...)

		// Deduplicate tags.
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
			// Log and skip invalid entries rather than aborting the whole import.
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

	color.Green("Added %d projects, updated %d projects from github.com/%s", added, updated, username)
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
