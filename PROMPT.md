**Here is a detailed, self-contained prompt** you can copy and paste directly to an AI coding assistant (Claude, Codex, etc.) to generate the complete Go CLI tool. The prompt incorporates every requirement from the conversation, including config-driven indexing, tag-based progressive disclosure, GitHub support (public/private/single-repo/user-level), local directories, ad-hoc additions, and a `showboat` subcommand that invokes Simon Willison’s local `showboat` binary on a project’s local directory.

---

**Prompt for AI Coder:**

You are an expert Go engineer. Build a complete, production-ready CLI tool in Go named `project-patterns` (binary name: `project-patterns`).

**Core Purpose**  
The tool maintains a personal index of reference projects (Git repositories and local directories) with tags and short descriptions. It enables progressive disclosure for AI coding agents: a user can run `project-patterns search --tags frontend elixir` and receive a clean list of matching projects with name, short description, and local path/URL. The user can then reference specific projects (“use the data-caching pattern from XYZ”) when prompting an AI.

**Key Features & Requirements**

1. **Configuration**  
   - Use Viper to load a YAML config file at `~/.config/project-patterns/config.yaml` (create directory and file if missing).  
   - Config structure must support:
     - Global settings (e.g., `cache_dir`, `github_token`, `showboat_path`).
     - A flat list of indexed projects under `projects:`.
     - Full example:
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
       ```

2. **Project Model** (each entry in the index)  
   ```yaml
   - id: unique-slug
     name: Human readable name
     type: local | git
     path: /absolute/local/path   # for local type
     url: https://github.com/user/repo   # for git type
     clone_path: ~/.cache/project-patterns/clones/user-repo   # managed local clone
     tags: [frontend, elixir, caching]
     description: "One-line description (max 120 chars). If missing, derive from README first paragraph or GitHub repo description."
     last_synced: "2026-04-03T10:00:00Z"
   ```

3. **Supported Addition Methods** (all via subcommands)
   - `add local <path> --name <name> --tags tag1,tag2 --desc "optional"`  
   - `add repo <git-url> [--name] [--tags] [--private]` (if private, use configured GitHub token for HTTPS clone with embedded token).  
   - `add github-user <username> [--include-private] [--tags default-tag1,default-tag2]` – uses GitHub API to list all repos for the user and adds them as individual projects (respect rate limits, handle pagination).  
   - `add dir <directory> [--recursive] [--tags ...]` – adds a top-level directory or every immediate sub-directory as a separate project (useful for monorepos or folders containing many small projects).  
   - All add commands append/update the config and auto-populate description when possible.

4. **Sync / Cloning Behavior**  
   - A `sync` command pulls/clones all git projects into `cache_dir` (default `~/.cache/project-patterns/clones`).  
   - Cloning uses `git clone` via `os/exec` (or `go-git` if you prefer). For private repos, construct HTTPS URL with embedded token (`https://<token>@github.com/...`).  
   - `showboat` and description extraction always operate on the local `clone_path` or `path`.

5. **Search / Progressive Disclosure**  
   - `search --tags tag1,tag2` (comma-separated, treated as AND).  
   - Output a clean table (use `github.com/jedib0t/go-pretty/v6/table`) showing: ID | Name | Description | Path | Tags.  
   - Support `--json` and `--markdown` flags for easy copying into AI prompts.  
   - Optional `--limit N` and `--query STRING` for text search on name/description (case-insensitive substring match). No additional dependencies for search.

6. **Showboat Integration**  
   - Subcommand `showboat <project-id-or-name>`  
   - Locates the project’s local path (`path` or `clone_path`).  
   - Executes the external `showboat` binary (path configurable in config, defaults to `showboat` in `$PATH`).  
   - Reasonable default invocation: `showboat summarize --dir <local-path> --output <cache-dir>/<id>-summary.md` (you may adjust the exact flags; document them clearly).  
   - If `showboat` is not found, print a helpful error with installation instructions (it is Simon Willison’s Go tool for generating executable Markdown demos).

7. **CLI Structure** (use Cobra)
   ```
   project-patterns
   ├── add (local | repo | github-user | dir)
   ├── search
   ├── sync
   ├── showboat <id>
   ├── list (alias for search with no tags)
   ├── config (show/edit)
   └── version
   ```
   Use standard Cobra best practices: persistent flags, help text, completion support (`cobra completion`).

8. **Technical Constraints & Best Practices**
   - Go 1.23+ module.  
   - Minimal external dependencies: `github.com/spf13/cobra`, `github.com/spf13/viper`, `github.com/jedib0t/go-pretty/v6/table`, `gopkg.in/yaml.v3`, `github.com/fatih/color`. Use only standard library for GitHub API (http.Client with token), git operations (os/exec), and other functionality.  
   - Graceful error handling and colored output.
   - Safe YAML read/write with backup on every write.  
   - Respect GitHub API rate limits (use `X-RateLimit-Remaining` headers).  
   - Idempotent: adding the same project again updates description/tags but does not duplicate.  
   - Comprehensive help text and examples in every command (exactly matching the use-case described in the original request).  
   - Include a `README.md` in the generated repository with installation, usage examples, and config template.

9. **Output**  
   Generate the full project as a single cohesive response containing:
   - `go.mod`  
   - `main.go` + all command files under `cmd/`  
   - `internal/config/config.go`, `internal/project/project.go`, `internal/github/client.go`  
   - Sample `config.yaml`  
   - Full `README.md` with usage examples, including the exact command `project-patterns search --tags frontend elixir`

Implement the tool exactly as specified. Prioritize clarity, robustness, and ease of use for an AI-agent workflow. You may adapt minor implementation details (such as exact helper functions, additional small packages if they simplify code without adding bloat, or slight adjustments to internal organization) using sound engineering judgment, as long as the configuration format, CLI interface, output formats, and core behaviors remain exactly as described. Do not add features outside this spec unless they directly support the described use-case.

---

Copy the entire block above and paste it into your coding agent. The resulting CLI will fully satisfy the requirements discussed in the Slack thread, including the `showboat` integration and the personal pattern-library workflow you already use. If you need any refinements to the prompt before using it, let me know.
