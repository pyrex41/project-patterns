package add

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pyrex41/project-patterns/internal/config"
	"github.com/pyrex41/project-patterns/internal/project"
	"github.com/spf13/cobra"
)

var dirCmd = &cobra.Command{
	Use:   "dir <directory>",
	Short: "Scan a directory and add all projects found within it",
	Long:  "Add every immediate subdirectory of <directory> as a separate local project. Useful for folders containing many small projects.",
	Example: `  project-patterns add dir ~/projects --tags personal
  project-patterns add dir ~/work/clients --recursive --tags work,client`,
	Args: cobra.ExactArgs(1),
	RunE: runDir,
}

func init() {
	dirCmd.Flags().Bool("recursive", false, "recursively scan nested subdirectories")
	dirCmd.Flags().StringSlice("tags", nil, "tags to apply to all discovered projects")
	AddCmd.AddCommand(dirCmd)
}

func runDir(cmd *cobra.Command, args []string) error {
	root := args[0]

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return fmt.Errorf("checking path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", absRoot)
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	dirs, err := findProjectDirs(absRoot, recursive)
	if err != nil {
		return fmt.Errorf("scanning directory: %w", err)
	}

	added, updated := 0, 0
	for _, d := range dirs {
		name := filepath.Base(d)
		id := project.GenerateID(name)

		desc := readReadmeFirstParagraph(d)

		p := project.Project{
			ID:          id,
			Name:        name,
			Type:        "local",
			Path:        d,
			Tags:        tags,
			Description: desc,
		}

		if cfg.AddOrUpdateProject(p) {
			updated++
		} else {
			added++
		}
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	color.Green("Added %d projects, updated %d projects from %s", added, updated, absRoot)
	return nil
}

func findProjectDirs(root string, recursive bool) ([]string, error) {
	if !recursive {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		var dirs []string
		for _, e := range entries {
			if e.IsDir() && e.Name()[0] != '.' {
				dirs = append(dirs, filepath.Join(root, e.Name()))
			}
		}
		return dirs, nil
	}

	var dirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !d.IsDir() || path == root {
			return nil
		}
		// Skip hidden directories.
		if d.Name()[0] == '.' {
			return filepath.SkipDir
		}
		dirs = append(dirs, path)
		return filepath.SkipDir // don't recurse into subdirs of project dirs
	})
	return dirs, err
}
