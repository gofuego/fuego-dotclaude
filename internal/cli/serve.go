package cli

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"

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
			"Path resolution matches `build`. Output and cache are written to a\n" +
			"scratch directory outside the content tree, so watching never loops.",
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

			// Build into a scratch dir outside the content tree. Serving in
			// place would write output/cache under the watched directory and
			// trigger an endless rebuild loop.
			work := serveWorkDir(res.ContentDir)

			eng := newEngine(res)
			return eng.Serve(context.Background(), engine.BuildOptions{
				ContentDir: res.ContentDir,
				OutputDir:  filepath.Join(work, "site"),
				CacheDir:   filepath.Join(work, "cache"),
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

// serveWorkDir returns a stable scratch directory (under the OS temp dir, keyed
// by the absolute content path) for a serve session's output and cache. Keeping
// it outside the content tree avoids a watch/rebuild loop; keying it by path lets
// the incremental cache survive restarts.
func serveWorkDir(contentDir string) string {
	abs, err := filepath.Abs(contentDir)
	if err != nil {
		abs = contentDir
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(abs))
	return filepath.Join(os.TempDir(), "fuego-dotclaude", fmt.Sprintf("%x", h.Sum64()))
}
