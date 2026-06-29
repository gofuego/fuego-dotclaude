package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gofuego/fuego-dotclaude/dotclaude"
	"github.com/gofuego/fuego/engine"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: "List the artifacts discovered in a .claude directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := resolveScope(cmd, args)
			if err != nil {
				return err
			}
			cfg, err := loadConfig(res.ContentDir)
			if err != nil {
				return err
			}

			cap := dotclaude.NewCapture()
			eng := newEngine(res)
			eng.Index(cap.Index())
			if _, err := eng.Validate(context.Background(), engine.BuildOptions{
				ContentDir: res.ContentDir,
				SiteName:   cfg.SiteName,
			}); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tNAME\tROUTE")
			for _, a := range dotclaude.ListArtifacts(cap.Pages()) {
				fmt.Fprintf(w, "%s\t%s\t%s\n", a.Type, a.Name, a.Route)
			}
			return w.Flush()
		},
	}
	addSiblingFlags(cmd)
	return cmd
}
