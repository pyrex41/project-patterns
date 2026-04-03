package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit hash, and build date of project-patterns.",
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.Bold)
		fmt.Fprintf(cmd.OutOrStdout(), "project-patterns %s (%s) built %s\n",
			bold.Sprint(Version), Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
