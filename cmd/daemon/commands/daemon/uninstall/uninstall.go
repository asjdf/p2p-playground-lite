package uninstall

import (
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/spf13/cobra"
	sysdaemon "github.com/takama/daemon"
)

// Cmd represents the uninstall command
var Cmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the daemon system service",
	Long:  `Remove the P2P Playground daemon from system services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := sysdaemon.New(consts.DaemonServiceName, consts.DaemonServiceDescription, sysdaemon.SystemDaemon)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		status, err := srv.Remove()
		if err != nil {
			return fmt.Errorf("failed to remove service: %w", err)
		}

		fmt.Println(status)
		return nil
	},
}
