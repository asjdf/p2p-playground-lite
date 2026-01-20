package status

import (
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/spf13/cobra"
	sysdaemon "github.com/takama/daemon"
)

// Cmd represents the status command
var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon service status",
	Long:  `Display the current status of the P2P Playground daemon system service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := sysdaemon.New(consts.DaemonServiceName, consts.DaemonServiceDescription, sysdaemon.SystemDaemon)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		status, err := srv.Status()
		if err != nil {
			fmt.Printf("Service status: %s\n", err)
			return nil
		}

		fmt.Println(status)
		return nil
	},
}
