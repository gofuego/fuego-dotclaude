package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gofuego/fuego-dotclaude/dotclaude"
	"github.com/gofuego/fuego-dotclaude/dotclaude/scope"
	dcconfig "github.com/gofuego/fuego-dotclaude/internal/config"
	"github.com/gofuego/fuego/engine"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var outputDir, baseURL string
	var incremental bool

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Build a documentation site from a .claude directory",
		Long: "Build a site from a .claude directory.\n\n" +
			"With no argument, builds ~/.claude in isolation. Given a repository\n" +
			"path that contains .claude/, builds it and folds in the root-level\n" +
			"CLAUDE.md/.mcp.json siblings. Given a .claude directory directly,\n" +
			"builds it in isolation.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := resolveScope(cmd, args)
			if err != nil {
				return err
			}

			cfg, err := loadConfig(res.ContentDir)
			if err != nil {
				return err
			}
			if outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			eng := newEngine(res)
			return eng.Build(context.Background(), engine.BuildOptions{
				ContentDir:  res.ContentDir,
				OutputDir:   cfg.OutputDir,
				SiteName:    cfg.SiteName,
				BaseURL:     cfg.BaseURL,
				Incremental: incremental,
			})
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (default: build)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL for deployment (e.g. /my-repo)")
	cmd.Flags().BoolVar(&incremental, "incremental", false, "reuse cached parses for unchanged files")
	addSiblingFlags(cmd)

	return cmd
}

// newEngine assembles the engine with the dotclaude pack, registering the
// sibling-injection hook when the resolution calls for it.
func newEngine(res scope.Resolution) *engine.Engine {
	eng := engine.New()
	eng.Use(dotclaude.Pack())
	if res.Siblings && res.SiblingDir != "" {
		eng.AfterParse(dotclaude.SiblingHook(res.SiblingDir))
	}
	return eng
}

// resolveScope maps the CLI argument and --siblings/--no-siblings flags to a
// scope.Resolution.
func resolveScope(cmd *cobra.Command, args []string) (scope.Resolution, error) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	home, _ := os.UserHomeDir()
	return scope.Resolve(scope.OSFS{}, arg, home, siblingsOverride(cmd))
}

// addSiblingFlags registers the mutually-informing siblings toggle.
func addSiblingFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("siblings", false, "force folding in root-level CLAUDE.md/.mcp.json")
	cmd.Flags().Bool("no-siblings", false, "suppress root-level sibling injection")
}

// siblingsOverride returns a pointer to the resolved override, or nil when
// neither flag was set (use the detected default).
func siblingsOverride(cmd *cobra.Command) *bool {
	switch {
	case cmd.Flags().Changed("no-siblings"):
		v := false
		return &v
	case cmd.Flags().Changed("siblings"):
		v := true
		return &v
	default:
		return nil
	}
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
