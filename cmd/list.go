package cmd

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all indexed projects",
	Long:  "List all projects in your index. Supports the same output formats as search.",
	Example: `  project-patterns list
  project-patterns list --json
  project-patterns list -v`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		mdOut, _ := cmd.Flags().GetBool("markdown")
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		verbose, _ := cmd.Flags().GetBool("verbose")
		doClone, _ := cmd.Flags().GetBool("clone")
		cloneDir, _ := cmd.Flags().GetString("clone-dir")
		cloneTarget := ""
		if cloneDir != "" {
			cloneTarget = cloneDir
		} else if doClone {
			cloneTarget = "__default__"
		}
		return filterAndDisplay(cmd, nil, query, limit, jsonOut, mdOut, verbose, cloneTarget)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("json", false, "Output as JSON")
	listCmd.Flags().Bool("markdown", false, "Output as Markdown table")
	listCmd.Flags().String("query", "", "Text search on name/description")
	listCmd.Flags().Int("limit", 0, "Limit number of results")
	listCmd.Flags().BoolP("verbose", "v", false, "Show full README for each project")
	listCmd.Flags().Bool("clone", false, "Clone matched git projects to their default cache paths")
	listCmd.Flags().String("clone-dir", "", "Clone matched git projects into this directory")
}
