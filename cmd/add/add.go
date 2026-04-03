package add

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add <path-or-url>",
	Short: "Add projects to the index",
	Long: `Add a project by local path or remote URL. The input type is auto-detected:

  Local path:              pp add ~/projects/myapp --tags go,web
  GitHub/GitLab repo:      pp add github.com/user/repo --tags tools
  GitHub/GitLab user:      pp add github.com/user --tags oss`,
	Args: cobra.ArbitraryArgs,
	RunE: runAdd,
}

func init() {
	AddCmd.Flags().String("name", "", "project name (defaults to directory/repo basename)")
	AddCmd.Flags().StringSlice("tags", nil, "tags to apply to the project")
	AddCmd.Flags().String("desc", "", "short description of the project")
	AddCmd.Flags().Bool("private", false, "mark project as private (repo only)")
	AddCmd.Flags().Bool("include-private", false, "include private repositories (user only, requires token)")
}

// knownHosts are forge hostnames we recognise for auto-detection.
var knownHosts = map[string]bool{
	"github.com": true,
	"gitlab.com": true,
}

type inputKind int

const (
	kindLocal inputKind = iota
	kindRepo
	kindUser
)

// classifyInput determines whether the argument is a local path, a remote
// repo URL, or a forge user/org reference.
func classifyInput(arg string) (inputKind, string) {
	// Full URL with scheme.
	if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "git@") {
		u, err := url.Parse(arg)
		if err != nil {
			return kindLocal, arg
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 1 && parts[0] != "" {
			return kindUser, arg
		}
		return kindRepo, arg
	}

	// Shorthand: github.com/user or gitlab.com/user/repo (no scheme).
	if i := strings.IndexByte(arg, '/'); i > 0 {
		host := arg[:i]
		if knownHosts[host] {
			rest := strings.Trim(arg[i+1:], "/")
			rest = strings.TrimSuffix(rest, ".git")
			parts := strings.Split(rest, "/")
			fullURL := "https://" + arg
			if len(parts) == 1 && parts[0] != "" {
				return kindUser, fullURL
			}
			if len(parts) >= 2 {
				return kindRepo, fullURL
			}
		}
	}

	// Check if it looks like a bare "host/path" for an unknown forge.
	// Fall through to local path.

	return kindLocal, arg
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	kind, resolved := classifyInput(args[0])

	// Propagate flags to the appropriate subcommand by re-invoking it.
	switch kind {
	case kindLocal:
		// Check if path exists as a directory — if not, show a helpful error.
		info, err := os.Stat(args[0])
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s\n(if this is a remote URL, prefix with https://)", args[0])
			}
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("not a directory: %s", args[0])
		}
		return runLocalWith(cmd, args[0])
	case kindRepo:
		return runRepoWith(cmd, resolved)
	case kindUser:
		return runGitHubUserWith(cmd, resolved)
	default:
		return fmt.Errorf("could not determine input type for %q", args[0])
	}
}
