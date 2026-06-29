package cli

import (
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Execute runs the fuego-dotclaude CLI.
func Execute() error {
	root := newRootCmd()
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fuego-dotclaude",
		Short:         "Render a .claude directory as a navigable documentation site",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		newBuildCmd(),
		newServeCmd(),
	)

	return cmd
}
