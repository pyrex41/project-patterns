package cmd

import (
	"fmt"
	"os"

	"github.com/pyrex41/project-patterns/cmd/add"
	"github.com/pyrex41/project-patterns/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	AppConfig *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "project-patterns",
	Short: "A personal index of reference projects for AI coding workflows",
	Long: `project-patterns is a CLI tool for indexing, searching, and syncing
reference projects used in AI-assisted coding workflows.

It lets you maintain a curated catalog of local directories and remote
repositories, tag them for retrieval, and pipe them into tools like showboat
for context injection.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for commands that don't need it.
		switch cmd.Name() {
		case "version", "help", "completion":
			return nil
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		AppConfig = cfg
		return nil
	},
}

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default: ~/.config/project-patterns/config.yaml)")
	rootCmd.AddCommand(add.AddCmd)
}
