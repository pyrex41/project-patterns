package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/pyrex41/project-patterns/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or edit configuration",
	Long:  "View and manage your project-patterns configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the current configuration",
	Long:  "Display the current configuration settings.",
	Example: `  project-patterns config show
  project-patterns config show --raw`,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")

		if raw {
			data, err := yaml.Marshal(AppConfig)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		}

		bold := color.New(color.Bold)
		cfgPath := cfgFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}

		bold.Fprintf(cmd.OutOrStdout(), "Config file: ")
		fmt.Fprintln(cmd.OutOrStdout(), cfgPath)

		bold.Fprintf(cmd.OutOrStdout(), "Cache dir:   ")
		fmt.Fprintln(cmd.OutOrStdout(), AppConfig.CacheDir)

		bold.Fprintf(cmd.OutOrStdout(), "GitHub token: ")
		if AppConfig.GitHubToken != "" {
			masked := strings.Repeat("*", 4)
			if len(AppConfig.GitHubToken) > 4 {
				masked += AppConfig.GitHubToken[len(AppConfig.GitHubToken)-4:]
			}
			fmt.Fprintln(cmd.OutOrStdout(), masked)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "not set")
		}

		bold.Fprintf(cmd.OutOrStdout(), "Showboat:    ")
		if AppConfig.ShowboatPath != "" {
			fmt.Fprintln(cmd.OutOrStdout(), AppConfig.ShowboatPath)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "default (PATH)")
		}

		bold.Fprintf(cmd.OutOrStdout(), "Projects:    ")
		fmt.Fprintf(cmd.OutOrStdout(), "%d indexed\n", len(AppConfig.Projects))

		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Open the configuration file in $EDITOR",
	Long:    "Open the configuration file in your default editor ($EDITOR).",
	Example: "  project-patterns config edit",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := cfgFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi"
		}

		c := exec.Command(editor, cfgPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configShowCmd.Flags().Bool("raw", false, "Output full config as raw YAML")
}
