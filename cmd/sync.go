package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/reubenb/project-patterns/internal/config"
	"github.com/reubenb/project-patterns/internal/git"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Clone or pull all git projects",
	Long:  "Clone or pull all git-type projects in your index. New projects are cloned to the cache directory; existing clones are updated with git pull.",
	Example: "  project-patterns sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := AppConfig

		var cloned, pulled, errors int

		for i := range cfg.Projects {
			p := &cfg.Projects[i]
			if p.Type != "git" {
				continue
			}

			dest := config.ExpandPath(p.ClonePath)
			if dest == "" {
				dest = config.ExpandPath(cfg.CacheDir) + "/" + p.ID
			}

			var opErr error
			if git.IsGitRepo(dest) {
				opErr = git.Pull(dest)
				if opErr != nil {
					color.Yellow("Warning: failed to pull %s: %v", p.Name, opErr)
					errors++
					continue
				}
				color.Green("Synced: %s", p.Name)
				pulled++
			} else {
				opErr = git.Clone(p.URL, dest, cfg.GitHubToken)
				if opErr != nil {
					color.Yellow("Warning: failed to clone %s: %v", p.Name, opErr)
					errors++
					continue
				}
				color.Green("Cloned: %s", p.Name)
				cloned++
			}

			p.ClonePath = dest
			p.LastSynced = time.Now().Format(time.RFC3339)
		}

		if err := cfg.Save(cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		total := cloned + pulled
		fmt.Printf("Synced %d projects (%d cloned, %d pulled, %d errors)\n", total, cloned, pulled, errors)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
