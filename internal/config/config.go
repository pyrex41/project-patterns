package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyrex41/project-patterns/internal/project"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration for project-patterns.
type Config struct {
	CacheDir     string            `yaml:"cache_dir"`
	GitHubToken  string            `yaml:"github_token,omitempty"`
	ShowboatPath string            `yaml:"showboat_path,omitempty"`
	Projects     []project.Project `yaml:"projects"`
}

// DefaultConfigDir returns the default configuration directory (~/.config/project-patterns).
func DefaultConfigDir() string {
	return ExpandPath("~/.config/project-patterns")
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// DefaultCacheDir returns the default cache directory (~/.cache/project-patterns/clones).
func DefaultCacheDir() string {
	return ExpandPath("~/.cache/project-patterns/clones")
}

// ExpandPath expands a leading ~ to the user's home directory.
func ExpandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return home + path[1:]
}

// Load reads the configuration from configPath using Viper. If configPath is
// empty, DefaultConfigPath() is used. The config directory and an empty default
// config file are created if they do not already exist.
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

	// Ensure the config directory exists.
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}

	// Create a default config file if none exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaults := &Config{
			CacheDir: DefaultCacheDir(),
			Projects: []project.Project{},
		}
		if writeErr := defaults.Save(configPath); writeErr != nil {
			return nil, fmt.Errorf("creating default config: %w", writeErr)
		}
	}

	v := viper.New()
	v.SetDefault("cache_dir", DefaultCacheDir())
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.CacheDir == "" {
		cfg.CacheDir = DefaultCacheDir()
	}

	return &cfg, nil
}

// Save marshals the Config to YAML and writes it to configPath. If configPath
// is empty, DefaultConfigPath() is used. An existing file is backed up to
// configPath+".bak" before writing. The write is performed atomically via a
// temporary file and os.Rename.
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

	// Ensure the config directory exists.
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Back up existing file if present.
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".bak"
		if err := copyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("backing up config: %w", err)
		}
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to a temp file in the same directory, then rename atomically.
	tmp, err := os.CreateTemp(configDir, ".config-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, configPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// FindProject returns the first project whose ID matches idOrName exactly, or
// whose Name matches case-insensitively. Returns nil if no match is found.
func (c *Config) FindProject(idOrName string) *project.Project {
	lower := strings.ToLower(idOrName)
	for i := range c.Projects {
		if c.Projects[i].ID == idOrName {
			return &c.Projects[i]
		}
		if strings.ToLower(c.Projects[i].Name) == lower {
			return &c.Projects[i]
		}
	}
	return nil
}

// AddOrUpdateProject adds p to the config if no project with the same ID
// exists, or replaces the existing entry if one does. Returns true if an
// existing project was updated, false if p was newly appended.
func (c *Config) AddOrUpdateProject(p project.Project) bool {
	for i := range c.Projects {
		if c.Projects[i].ID == p.ID {
			c.Projects[i] = p
			return true
		}
	}
	c.Projects = append(c.Projects, p)
	return false
}

// copyFile copies src to dst, creating or truncating dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
