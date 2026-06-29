// Package config handles the optional fuego-dotclaude.yaml override file. It
// holds only the site-level settings the CLI feeds to the engine; the routes,
// theme, and parser all come from the dotclaude pack.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the optional fuego-dotclaude.yaml override file.
type Config struct {
	SiteName  string `yaml:"site_name"`
	BaseURL   string `yaml:"base_url"`
	OutputDir string `yaml:"output_path"`
}

// Defaults returns a Config with all default values.
func Defaults() *Config {
	return &Config{
		SiteName:  "Claude Code Workspace",
		BaseURL:   "",
		OutputDir: "build",
	}
}

// Load reads a fuego-dotclaude.yaml file and merges it with defaults. If the
// file doesn't exist, defaults are returned.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	override := &Config{}
	if err := yaml.Unmarshal(data, override); err != nil {
		return nil, err
	}

	merge(cfg, override)
	return cfg, nil
}

// merge applies non-zero override values onto the base config.
func merge(base, override *Config) {
	if override.SiteName != "" {
		base.SiteName = override.SiteName
	}
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.OutputDir != "" {
		base.OutputDir = override.OutputDir
	}
}
