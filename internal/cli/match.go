package cli

import (
	"github.com/bssm-oss/chess-wifi/internal/tui"
	"github.com/spf13/cobra"
)

func newMatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "match",
		Short: "Host or join a LAN chess match",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}

	return cmd
}
