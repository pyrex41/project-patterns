package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pyrex41/project-patterns/internal/config"
	"github.com/spf13/cobra"
)

var showboatCmd = &cobra.Command{
	Use:   "showboat <project-id-or-name>",
	Short: "Run showboat on a project",
	Long:  "Run the showboat tool on a project's local directory to generate a summary.",
	Example: `  project-patterns showboat my-project
  project-patterns showboat "Elixir Data Caching Pattern"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := AppConfig.FindProject(args[0])
		if p == nil {
			return fmt.Errorf("project not found: %s", args[0])
		}

		localPath := p.LocalPath()
		if localPath == "" {
			return fmt.Errorf("project %q has no local path — run 'project-patterns sync' first", p.Name)
		}

		// Determine showboat binary.
		sbPath := AppConfig.ShowboatPath
		if sbPath == "" {
			sbPath = "showboat"
		}

		resolved, err := exec.LookPath(sbPath)
		if err != nil {
			color.Red("Error: showboat binary not found at %q\n", sbPath)
			fmt.Println()
			fmt.Println("showboat is a tool by Simon Willison for generating executable Markdown demos.")
			fmt.Println()
			fmt.Println("To install:")
			fmt.Println("  go install github.com/simonw/showboat@latest")
			fmt.Println()
			fmt.Println("Or set the path in your config:")
			fmt.Println("  showboat_path: /path/to/showboat")
			return fmt.Errorf("showboat not found")
		}

		cacheDir := config.ExpandPath(AppConfig.CacheDir)
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			return fmt.Errorf("creating cache directory: %w", err)
		}

		outputPath := filepath.Join(cacheDir, p.ID+"-summary.md")

		c := exec.Command(resolved, "summarize", "--dir", localPath, "--output", outputPath)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return fmt.Errorf("showboat failed: %w", err)
		}

		color.Green("Summary written to: %s", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showboatCmd)
}
