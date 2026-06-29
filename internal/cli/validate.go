package cli

import (
	"context"
	"fmt"

	"github.com/gofuego/fuego-dotclaude/dotclaude"
	"github.com/gofuego/fuego/engine"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Check a .claude directory for coherence problems",
		Long: "Report advisory diagnostics (missing frontmatter, skill name/dir\n" +
			"mismatches, dangling references, malformed JSON). Warnings by default;\n" +
			"--strict exits non-zero so CI can gate on a clean tree.",
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

			cap := dotclaude.NewCapture()
			eng := newEngine(res)
			eng.Index(cap.Index())
			if _, err := eng.Validate(context.Background(), engine.BuildOptions{
				ContentDir: res.ContentDir,
				SiteName:   cfg.SiteName,
			}); err != nil {
				return err
			}

			diags := dotclaude.Diagnostics(cap.Pages())
			for _, d := range diags {
				fmt.Printf("%s: %s\n", d.RelPath, d.Message)
			}
			if len(diags) == 0 {
				fmt.Println("no problems found")
				return nil
			}
			fmt.Printf("%d problem(s) found\n", len(diags))
			if strict {
				return fmt.Errorf("validation failed in --strict mode")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "exit non-zero if any problem is found")
	addSiblingFlags(cmd)
	return cmd
}
