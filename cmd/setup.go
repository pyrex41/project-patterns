package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install pp alias and shell completions",
	Long: `Set up the "pp" alias and shell completions for your current shell.

Detects fish, zsh, or bash and writes the appropriate config. Run this
once after installing project-patterns.`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().Bool("no-alias", false, "skip installing the pp alias")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	noAlias, _ := cmd.Flags().GetBool("no-alias")
	shell := detectShell()

	switch shell {
	case "fish":
		return setupFish(cmd, noAlias)
	case "zsh":
		return setupZsh(cmd, noAlias)
	case "bash":
		return setupBash(cmd, noAlias)
	default:
		return fmt.Errorf("unsupported shell %q — please set up manually", shell)
	}
}

func detectShell() string {
	// Check SHELL env var.
	sh := os.Getenv("SHELL")
	if strings.Contains(sh, "fish") {
		return "fish"
	}
	if strings.Contains(sh, "zsh") {
		return "zsh"
	}
	if strings.Contains(sh, "bash") {
		return "bash"
	}
	return filepath.Base(sh)
}

func setupFish(cmd *cobra.Command, noAlias bool) error {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}

	// Install completions.
	compDir := filepath.Join(configDir, "fish", "completions")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		return fmt.Errorf("creating completions dir: %w", err)
	}

	compFile := filepath.Join(compDir, "project-patterns.fish")
	compOut, err := generateCompletion(cmd, "fish")
	if err != nil {
		return err
	}
	if err := os.WriteFile(compFile, []byte(compOut), 0o644); err != nil {
		return fmt.Errorf("writing completions: %w", err)
	}
	color.Green("Installed completions: %s", compFile)

	if !noAlias {
		// Write pp alias + completion wrapper.
		ppComp := filepath.Join(compDir, "pp.fish")
		if err := os.WriteFile(ppComp, []byte("complete -c pp -w project-patterns\n"), 0o644); err != nil {
			return fmt.Errorf("writing pp completions: %w", err)
		}

		// Add alias to config.fish.
		confFile := filepath.Join(configDir, "fish", "config.fish")
		aliasLine := "alias pp='project-patterns'"
		if !fileContains(confFile, aliasLine) {
			f, err := os.OpenFile(confFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return fmt.Errorf("opening config.fish: %w", err)
			}
			fmt.Fprintf(f, "\n%s\n", aliasLine)
			f.Close()
			color.Green("Added alias to %s", confFile)
		} else {
			color.Yellow("Alias already in %s", confFile)
		}
		color.Green("Installed pp completions: %s", ppComp)
	}

	fmt.Println("\nRestart your shell or run: source", filepath.Join(configDir, "fish", "config.fish"))
	return nil
}

func setupZsh(cmd *cobra.Command, noAlias bool) error {
	home, _ := os.UserHomeDir()

	// Install completions.
	compDir := filepath.Join(home, ".zsh", "completions")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		return fmt.Errorf("creating completions dir: %w", err)
	}

	compFile := filepath.Join(compDir, "_project-patterns")
	compOut, err := generateCompletion(cmd, "zsh")
	if err != nil {
		return err
	}
	if err := os.WriteFile(compFile, []byte(compOut), 0o644); err != nil {
		return fmt.Errorf("writing completions: %w", err)
	}
	color.Green("Installed completions: %s", compFile)

	rcFile := filepath.Join(home, ".zshrc")

	// Ensure completions dir is in fpath.
	fpathLine := fmt.Sprintf("fpath=(%s $fpath)", compDir)
	if !fileContains(rcFile, compDir) {
		appendToFile(rcFile, fpathLine)
		color.Green("Added fpath to %s", rcFile)
	}

	if !noAlias {
		aliasLine := "alias pp='project-patterns'"
		if !fileContains(rcFile, aliasLine) {
			appendToFile(rcFile, aliasLine)
			color.Green("Added alias to %s", rcFile)
		} else {
			color.Yellow("Alias already in %s", rcFile)
		}

		// Also install a compdef for pp.
		compdefLine := "compdef pp=project-patterns"
		if !fileContains(rcFile, compdefLine) {
			appendToFile(rcFile, compdefLine)
		}
	}

	fmt.Println("\nRestart your shell or run: source", rcFile)
	return nil
}

func setupBash(cmd *cobra.Command, noAlias bool) error {
	home, _ := os.UserHomeDir()

	// Install completions.
	compDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		return fmt.Errorf("creating completions dir: %w", err)
	}

	compFile := filepath.Join(compDir, "project-patterns")
	compOut, err := generateCompletion(cmd, "bash")
	if err != nil {
		return err
	}
	if err := os.WriteFile(compFile, []byte(compOut), 0o644); err != nil {
		return fmt.Errorf("writing completions: %w", err)
	}
	color.Green("Installed completions: %s", compFile)

	rcFile := filepath.Join(home, ".bashrc")

	if !noAlias {
		aliasLine := "alias pp='project-patterns'"
		if !fileContains(rcFile, aliasLine) {
			appendToFile(rcFile, aliasLine)
			color.Green("Added alias to %s", rcFile)
		} else {
			color.Yellow("Alias already in %s", rcFile)
		}

		// Make completions work for the alias too.
		completeWrap := "complete -o default -F __start_project-patterns pp"
		if !fileContains(rcFile, completeWrap) {
			appendToFile(rcFile, completeWrap)
		}
	}

	fmt.Println("\nRestart your shell or run: source", rcFile)
	return nil
}

func generateCompletion(cmd *cobra.Command, shell string) (string, error) {
	// Use the actual binary to generate completions so they're always current.
	bin, err := exec.LookPath("project-patterns")
	if err != nil {
		// Fall back to generating from the current cobra tree.
		bin = os.Args[0]
	}
	out, err := exec.Command(bin, "completion", shell).Output()
	if err != nil {
		return "", fmt.Errorf("generating %s completions: %w", shell, err)
	}
	return string(out), nil
}

func fileContains(path, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

func appendToFile(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s\n", line)
	return err
}
