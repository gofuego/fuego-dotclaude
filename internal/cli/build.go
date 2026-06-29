package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gofuego/fuego-dotclaude/dotclaude"
	dcconfig "github.com/gofuego/fuego-dotclaude/internal/config"
	"github.com/gofuego/fuego/engine"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var outputDir string
	var baseURL string
	var incremental bool

	cmd := &cobra.Command{
		Use:   "build [claude-dir]",
		Short: "Build a documentation site from a .claude directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			claudeDir := "."
			if len(args) > 0 {
				claudeDir = args[0]
			}

			cfg, err := loadConfig(claudeDir)
			if err != nil {
				return err
			}
			if outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			return buildSite(claudeDir, cfg, incremental)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (default: build)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL for deployment (e.g. /my-repo)")
	cmd.Flags().BoolVar(&incremental, "incremental", false, "reuse cached parses for unchanged files")

	return cmd
}

// buildSite assembles the engine with the dotclaude pack and runs a build. The
// pack supplies the parser, theme, routes, and hooks; the CLI supplies only the
// site-specific dirs and metadata.
func buildSite(claudeDir string, cfg *dcconfig.Config, incremental bool) error {
	eng := engine.New()
	eng.Use(dotclaude.Pack())
	return eng.Build(context.Background(), engine.BuildOptions{
		ContentDir:  claudeDir,
		OutputDir:   cfg.OutputDir,
		SiteName:    cfg.SiteName,
		BaseURL:     cfg.BaseURL,
		Incremental: incremental,
	})
}

// loadConfig finds and loads fuego-dotclaude.yaml from the content directory,
// its parent, or the working directory.
func loadConfig(claudeDir string) (*dcconfig.Config, error) {
	candidates := []string{
		filepath.Join(claudeDir, "fuego-dotclaude.yaml"),
		filepath.Join(claudeDir, "..", "fuego-dotclaude.yaml"),
		"fuego-dotclaude.yaml",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return dcconfig.Load(path)
		}
	}

	return dcconfig.Defaults(), nil
}
