package cli

import (
	"context"

	"github.com/gofuego/fuego/engine"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var baseURL string
	var port int

	cmd := &cobra.Command{
		Use:   "serve [path]",
		Short: "Serve a .claude site with live reload",
		Long: "Build and serve a .claude directory, rebuilding on change.\n\n" +
			"Path resolution matches `build`. Live-reload covers every file inside\n" +
			"the .claude directory.",
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
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			eng := newEngine(res)
			return eng.Serve(context.Background(), engine.BuildOptions{
				ContentDir: res.ContentDir,
				OutputDir:  cfg.OutputDir,
				SiteName:   cfg.SiteName,
				BaseURL:    cfg.BaseURL,
				DevPort:    port,
			})
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL for deployment (e.g. /my-repo)")
	cmd.Flags().IntVar(&port, "port", 0, "dev server port (default: engine default)")
	addSiblingFlags(cmd)

	return cmd
}
