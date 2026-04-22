package cli

import "github.com/spf13/cobra"

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "chess-wifi",
		Short:         "Play peer-to-peer LAN chess in a polished terminal UI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newMatchCommand())
	return cmd
}
