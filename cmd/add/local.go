package add

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenb/project-patterns/internal/config"
	"github.com/reubenb/project-patterns/internal/project"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local <path>",
	Short: "Add a local directory as a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runLocal,
}

func init() {
	localCmd.Flags().String("name", "", "project name (defaults to directory basename)")
	localCmd.Flags().StringSlice("tags", nil, "tags to apply to the project")
	localCmd.Flags().String("desc", "", "short description of the project")
	AddCmd.AddCommand(localCmd)
}

func runLocal(cmd *cobra.Command, args []string) error {
	dirArg := args[0]

	// Resolve to absolute path.
	absPath, err := filepath.Abs(dirArg)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Validate the path exists and is a directory.
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
		return fmt.Errorf("checking path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Determine name.
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = filepath.Base(absPath)
	}

	tags, _ := cmd.Flags().GetStringSlice("tags")
	desc, _ := cmd.Flags().GetString("desc")

	// Try to read description from README if not provided.
	if desc == "" {
		desc = readReadmeFirstParagraph(absPath)
	}

	id := project.GenerateID(name)

	p := project.Project{
		ID:          id,
		Name:        name,
		Type:        "local",
		Path:        absPath,
		Tags:        tags,
		Description: desc,
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	updated := cfg.AddOrUpdateProject(p)

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if updated {
		color.Yellow("Updated project: %s (%s)", name, id)
	} else {
		color.Green("Added project: %s (%s)", name, id)
	}

	return nil
}

// readReadmeFirstParagraph looks for a README file in dir and returns the first
// non-empty paragraph of plain text (skipping Markdown heading lines).
func readReadmeFirstParagraph(dir string) string {
	candidates := []string{"README.md", "README.MD", "README.txt", "README"}
	for _, name := range candidates {
		p := filepath.Join(dir, name)
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		var lines []string
		scanner := bufio.NewScanner(f)
		inParagraph := false
		for scanner.Scan() {
			line := scanner.Text()
			// Skip Markdown headings and horizontal rules.
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "===") {
				if inParagraph {
					break
				}
				continue
			}
			if trimmed == "" {
				if inParagraph {
					break
				}
				continue
			}
			inParagraph = true
			lines = append(lines, trimmed)
		}
		if len(lines) > 0 {
			return strings.Join(lines, " ")
		}
	}
	return ""
}
