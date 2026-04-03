package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/reubenb/project-patterns/internal/config"
	"github.com/reubenb/project-patterns/internal/git"
	gh "github.com/reubenb/project-patterns/internal/github"
	"github.com/reubenb/project-patterns/internal/project"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search projects by tags and keywords",
	Long: `Search your indexed projects by tags and/or keywords.

Tags are matched with AND logic — a project must have all specified tags.
Text queries match case-insensitively against project name and description.

Use -v/--verbose to display the full README for each matching project.
Use --clone to clone all matched git projects locally.`,
	Example: `  project-patterns search --tags frontend,elixir
  project-patterns search --query "data caching" --json
  project-patterns search --tags backend --limit 5 --markdown
  project-patterns search --tags elixir -v
  project-patterns search --tags elixir --clone
  project-patterns search --tags elixir --clone-dir ~/my-refs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, _ := cmd.Flags().GetStringSlice("tags")
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOut, _ := cmd.Flags().GetBool("json")
		mdOut, _ := cmd.Flags().GetBool("markdown")
		verbose, _ := cmd.Flags().GetBool("verbose")
		doClone, _ := cmd.Flags().GetBool("clone")
		cloneDir, _ := cmd.Flags().GetString("clone-dir")
		cloneTarget := ""
		if cloneDir != "" {
			cloneTarget = cloneDir
		} else if doClone {
			cloneTarget = "__default__"
		}
		return filterAndDisplay(cmd, tags, query, limit, jsonOut, mdOut, verbose, cloneTarget)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSlice("tags", nil, "Filter by tags (AND logic)")
	searchCmd.Flags().String("query", "", "Text search on name/description")
	searchCmd.Flags().Int("limit", 0, "Limit number of results")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("markdown", false, "Output as Markdown table")
	searchCmd.Flags().BoolP("verbose", "v", false, "Show full README for each project")
	searchCmd.Flags().Bool("clone", false, "Clone matched git projects to their default cache paths")
	searchCmd.Flags().String("clone-dir", "", "Clone matched git projects into this directory")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// fetchReadme returns the full README content for a project.
// Checks local path first, falls back to GitHub API for git projects.
func fetchReadme(p *project.Project) string {
	// Try local path first (works for local projects and cloned git projects).
	localPath := p.LocalPath()
	if localPath != "" {
		for _, name := range []string{"README.md", "README.MD", "README.txt", "README"} {
			data, err := os.ReadFile(filepath.Join(localPath, name))
			if err == nil {
				return string(data)
			}
		}
	}

	// For git projects, fetch from GitHub API.
	if p.Type == "git" && p.URL != "" {
		owner, repo := parseGitHubOwnerRepo(p.URL)
		if owner != "" && repo != "" {
			token := ""
			if AppConfig != nil {
				token = AppConfig.GitHubToken
			}
			client := gh.NewClient(token)
			content := client.FetchReadme(owner, repo)
			if content != "" {
				return content
			}
		}
	}

	return ""
}

// parseGitHubOwnerRepo extracts owner and repo from a GitHub URL.
func parseGitHubOwnerRepo(rawURL string) (string, string) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host != "github.com" {
		return "", ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git")
}

type projectResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Path        string   `json:"path,omitempty"`
	URL         string   `json:"url,omitempty"`
	ClonePath   string   `json:"clone_path,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags"`
	LastSynced  string   `json:"last_synced,omitempty"`
	Readme      string   `json:"readme,omitempty"`
}

type displayRow struct {
	id          string
	name        string
	description string
	path        string
	tags        string
}

func filterAndDisplay(cmd *cobra.Command, tags []string, query string, limit int, jsonOut, mdOut, verbose bool, cloneDir string) error {
	var jsonResults []projectResult
	var rows []displayRow
	var matched []project.Project
	// matchedIdx tracks which index in AppConfig.Projects each matched project came from.
	var matchedIdx []int

	projects := AppConfig.Projects

	for i, p := range projects {
		if len(tags) > 0 {
			var cleanTags []string
			for _, t := range tags {
				if t != "" {
					cleanTags = append(cleanTags, t)
				}
			}
			if len(cleanTags) > 0 && !p.HasAllTags(cleanTags) {
				continue
			}
		}
		if query != "" && !p.MatchesQuery(query) {
			continue
		}

		localPath := p.LocalPath()
		displayPath := localPath
		if displayPath == "" {
			displayPath = p.URL
		}

		pr := projectResult{
			ID:          p.ID,
			Name:        p.Name,
			Type:        p.Type,
			Path:        p.Path,
			URL:         p.URL,
			ClonePath:   p.ClonePath,
			Description: p.Description,
			Tags:        p.Tags,
			LastSynced:  p.LastSynced,
		}

		jsonResults = append(jsonResults, pr)
		matched = append(matched, p)
		matchedIdx = append(matchedIdx, i)
		rows = append(rows, displayRow{
			id:          p.ID,
			name:        p.Name,
			description: truncate(p.Description, 50),
			path:        displayPath,
			tags:        strings.Join(p.Tags, ", "),
		})
	}

	// Apply limit.
	if limit > 0 && limit < len(jsonResults) {
		jsonResults = jsonResults[:limit]
		rows = rows[:limit]
		matched = matched[:limit]
		matchedIdx = matchedIdx[:limit]
	}

	// Clone matched git projects if requested.
	if cloneDir != "" {
		cloned, pulled, errors := cloneMatched(matched, matchedIdx, cloneDir)
		// Update rows/jsonResults with new local paths after cloning.
		for i, idx := range matchedIdx {
			p := &AppConfig.Projects[idx]
			localPath := p.LocalPath()
			if localPath != "" {
				rows[i].path = localPath
				jsonResults[i].ClonePath = p.ClonePath
				jsonResults[i].Path = p.Path
			}
		}
		total := cloned + pulled
		if total > 0 || errors > 0 {
			color.Green("Cloned %d, pulled %d, %d errors", cloned, pulled, errors)
			fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	// Fetch READMEs if verbose.
	if verbose {
		for i := range matched {
			readme := fetchReadme(&matched[i])
			jsonResults[i].Readme = readme
		}
	}

	// JSON output.
	if jsonOut {
		out, err := json.MarshalIndent(jsonResults, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}

	// Markdown output.
	if mdOut {
		fmt.Fprintln(cmd.OutOrStdout(), "| ID | Name | Description | Path | Tags |")
		fmt.Fprintln(cmd.OutOrStdout(), "|----|------|-------------|------|------|")
		for i, r := range rows {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %s | %s | %s |\n",
				r.id, r.name, r.description, r.path, r.tags)
			if verbose && jsonResults[i].Readme != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "\n<details><summary>README: %s</summary>\n\n%s\n\n</details>\n\n",
					r.name, jsonResults[i].Readme)
			}
		}
		return nil
	}

	// Table output (default).
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No projects found.")
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"ID", "Name", "Description", "Path", "Tags"})
	for _, r := range rows {
		t.AppendRow(table.Row{r.id, r.name, r.description, r.path, r.tags})
	}
	t.Render()

	// Print READMEs after the table when verbose.
	if verbose {
		bold := color.New(color.Bold)
		for i, r := range rows {
			readme := jsonResults[i].Readme
			if readme == "" {
				continue
			}
			fmt.Fprintln(cmd.OutOrStdout())
			bold.Fprintf(cmd.OutOrStdout(), "─── README: %s ", r.name)
			fmt.Fprintln(cmd.OutOrStdout(), "───────────────────────────────────────")
			fmt.Fprintln(cmd.OutOrStdout(), readme)
		}
	}

	return nil
}

// cloneMatched clones or pulls all git-type projects from the matched set.
// If cloneDir is "__default__", uses each project's configured clone_path
// (or the global cache_dir). Otherwise, clones all into the specified directory.
// Updates AppConfig.Projects in place and saves config.
func cloneMatched(matched []project.Project, matchedIdx []int, cloneDir string) (cloned, pulled, errors int) {
	useCustomDir := cloneDir != "__default__"
	var baseDir string
	if useCustomDir {
		abs, err := filepath.Abs(cloneDir)
		if err != nil {
			color.Red("Error resolving clone path: %v", err)
			return 0, 0, 1
		}
		baseDir = abs
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			color.Red("Error creating clone directory: %v", err)
			return 0, 0, 1
		}
	}

	configDirty := false

	for i, p := range matched {
		if p.Type != "git" || p.URL == "" {
			continue
		}

		idx := matchedIdx[i]
		cp := &AppConfig.Projects[idx]

		// Determine destination.
		var dest string
		if useCustomDir {
			dest = filepath.Join(baseDir, p.ID)
		} else if cp.ClonePath != "" {
			dest = config.ExpandPath(cp.ClonePath)
		} else {
			dest = filepath.Join(config.ExpandPath(AppConfig.CacheDir), p.ID)
		}

		if git.IsGitRepo(dest) {
			if err := git.Pull(dest); err != nil {
				color.Yellow("  Warning: pull failed for %s: %v", p.Name, err)
				errors++
				continue
			}
			color.Green("  Pulled: %s → %s", p.Name, dest)
			pulled++
		} else {
			if err := git.Clone(p.URL, dest, AppConfig.GitHubToken); err != nil {
				color.Yellow("  Warning: clone failed for %s: %v", p.Name, err)
				errors++
				continue
			}
			color.Green("  Cloned: %s → %s", p.Name, dest)
			cloned++
		}

		cp.ClonePath = dest
		cp.LastSynced = time.Now().Format(time.RFC3339)
		// Update the local matched copy too so verbose/display sees the new path.
		matched[i].ClonePath = dest
		configDirty = true
	}

	if configDirty {
		if err := AppConfig.Save(cfgFile); err != nil {
			color.Red("Error saving config: %v", err)
		}
	}

	return
}
