package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search projects by tags and keywords",
	Long: `Search your indexed projects by tags and/or keywords.

Tags are matched with AND logic — a project must have all specified tags.
Text queries match case-insensitively against project name and description.`,
	Example: `  project-patterns search --tags frontend,elixir
  project-patterns search --query "data caching" --json
  project-patterns search --tags backend --limit 5 --markdown`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, err := cmd.Flags().GetStringSlice("tags")
		if err != nil {
			return err
		}
		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return err
		}
		jsonOut, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		mdOut, err := cmd.Flags().GetBool("markdown")
		if err != nil {
			return err
		}
		return filterAndDisplay(cmd, tags, query, limit, jsonOut, mdOut)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSlice("tags", nil, "Filter by tags (AND logic)")
	searchCmd.Flags().String("query", "", "Text search on name/description")
	searchCmd.Flags().Int("limit", 0, "Limit number of results")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("markdown", false, "Output as Markdown table")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func filterAndDisplay(cmd *cobra.Command, tags []string, query string, limit int, jsonOut, mdOut bool) error {
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
	}

	type displayRow struct {
		id          string
		name        string
		description string
		path        string
		tags        string
	}

	var jsonResults []projectResult
	var rows []displayRow

	projects := AppConfig.Projects

	// Filter
	for _, p := range projects {
		p := p // capture
		if len(tags) > 0 {
			// Remove empty strings that StringSlice may produce
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

		jsonResults = append(jsonResults, projectResult{
			ID:          p.ID,
			Name:        p.Name,
			Type:        p.Type,
			Path:        p.Path,
			URL:         p.URL,
			ClonePath:   p.ClonePath,
			Description: p.Description,
			Tags:        p.Tags,
			LastSynced:  p.LastSynced,
		})
		rows = append(rows, displayRow{
			id:          p.ID,
			name:        p.Name,
			description: truncate(p.Description, 50),
			path:        displayPath,
			tags:        strings.Join(p.Tags, ", "),
		})
	}

	// Apply limit
	if limit > 0 {
		if limit < len(jsonResults) {
			jsonResults = jsonResults[:limit]
			rows = rows[:limit]
		}
	}

	// Output
	if jsonOut {
		out, err := json.MarshalIndent(jsonResults, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}

	if mdOut {
		fmt.Fprintln(cmd.OutOrStdout(), "| ID | Name | Description | Path | Tags |")
		fmt.Fprintln(cmd.OutOrStdout(), "|----|------|-------------|------|------|")
		for _, r := range rows {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %s | %s | %s |\n",
				r.id, r.name, r.description, r.path, r.tags)
		}
		return nil
	}

	// Table output (default)
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

	return nil
}
