# project-patterns

A CLI tool for maintaining a personal index of reference projects — Git repositories and local directories — with tags and descriptions. Designed for progressive disclosure in AI coding workflows: search by tags, get clean structured output, and reference specific patterns when prompting.

## Installation

```bash
go install github.com/pyrex41/project-patterns@latest
```

Then run setup to install the `pp` alias and shell completions:

```bash
project-patterns setup
```

Or build from source:

```bash
git clone https://github.com/pyrex41/project-patterns.git
cd project-patterns
go build -o project-patterns .
./project-patterns setup
```

## Quick Start

```bash
# Add a local project (auto-detected from path)
project-patterns add ~/projects/my-app --tags backend,go --desc "REST API with caching"

# Add a remote repository (auto-detected from URL, description fetched from README)
project-patterns add github.com/user/repo --tags frontend,react
project-patterns add gitlab.com/org/project --tags work

# Bulk-add all repos for a GitHub user (auto-detected: host/user with no repo)
project-patterns add github.com/myusername --tags personal

# Add all projects in a directory
project-patterns add dir ~/projects --tags work

# Subcommands still work explicitly
project-patterns add local ~/projects/my-app --tags go
project-patterns add repo https://github.com/user/repo --tags react
project-patterns add github-user myusername --tags personal

# Search by tags
project-patterns search --tags frontend,elixir

# Search with text query
project-patterns search --query "data caching" --json

# Search with full README output
project-patterns search --tags elixir -v

# Search and clone matched projects locally
project-patterns search --tags elixir --clone
project-patterns search --tags elixir --clone-dir ~/my-refs

# List everything
project-patterns list --markdown

# Sync (clone/pull) all git projects
project-patterns sync

# Run showboat on a project
project-patterns showboat my-project
```

## Commands

| Command | Description |
|---------|-------------|
| `add <path-or-url>` | Auto-detect and add a local path, repo URL, or user |
| `add local <path>` | Add a local directory as a project |
| `add repo <url>` | Add a remote git repository |
| `add github-user <username>` | Bulk-add repos for a GitHub user |
| `add dir <directory>` | Add subdirectories as projects |
| `search` | Search by tags and/or keywords |
| `list` | List all indexed projects |
| `sync` | Clone or pull all git projects |
| `showboat <id>` | Run showboat on a project |
| `setup` | Install pp alias and shell completions |
| `config show` | Display current configuration |
| `config edit` | Open config in $EDITOR |
| `version` | Print version information |

## Configuration

Config file: `~/.config/project-patterns/config.yaml`

```yaml
cache_dir: ~/.cache/project-patterns/clones
github_token: ghp_...
showboat_path: /usr/local/bin/showboat
projects:
  - id: data-caching
    name: Elixir Data Caching Pattern
    type: git
    url: https://github.com/user/repo
    clone_path: ~/.cache/project-patterns/clones/user-repo
    tags: [elixir, caching, backend]
    description: "Uses ETS and GenServer for efficient caching with TTL."
    last_synced: "2026-04-03T10:00:00Z"
  - id: my-app
    name: my-app
    type: local
    path: /home/user/projects/my-app
    tags: [go, backend]
    description: "REST API with PostgreSQL and Redis."
```

## Output Formats

Search and list support three output formats:

- **Table** (default): Clean ASCII table via go-pretty
- **JSON** (`--json`): Machine-readable, ideal for piping into AI tools
- **Markdown** (`--markdown`): Pipe-delimited table for pasting into prompts

Add `-v`/`--verbose` to any format to include the full README for each project (fetched from disk or GitHub API).

## Cloning on Search

Use `--clone` or `--clone-dir <path>` with `search` or `list` to clone all matched git projects:

```bash
# Clone to default cache paths
project-patterns search --tags backend --clone

# Clone into a specific directory
project-patterns search --tags backend --clone-dir ~/refs
```

Already-cloned repos are pulled instead. The path column updates to show the local directory.

## Shell Setup

`project-patterns setup` auto-detects your shell (fish/zsh/bash) and installs:
- The `pp` alias
- Shell completions for both `project-patterns` and `pp`

```bash
project-patterns setup              # install alias + completions
project-patterns setup --no-alias   # completions only, no pp alias
```

Or install completions manually:

```bash
project-patterns completion fish > ~/.config/fish/completions/project-patterns.fish
project-patterns completion zsh > "${fpath[1]}/_project-patterns"
project-patterns completion bash > /etc/bash_completion.d/project-patterns
```

## How This Was Built

This project was built entirely through a multi-agent AI pipeline:

1. **Spec generation** — The initial `PROMPT.md` was drafted conversationally on [grok.com](https://grok.com), capturing the full CLI contract, config schema, and feature requirements.
2. **Task decomposition** — [`scud generate`](https://github.com/pyrex41/scud) parsed the prompt into a dependency-aware task graph (`.scud/tasks/tasks.scg`), breaking the spec into ~19 sub-tasks with agent assignments and complexity ratings.
3. **Plan review** — `scud heavy` ran a multi-agent review panel against both the prompt and task graph, synthesizing feedback from five specialist perspectives to simplify the spec and align it with "simple, flexible, and just works."
4. **Implementation** — Claude Code (Opus) designed a 5-wave execution plan and dispatched teams of Sonnet agents working in parallel, each owning disjoint file sets to avoid merge conflicts. The waves built up from foundation packages → infrastructure → command implementations → polish.

No code was written by hand.
