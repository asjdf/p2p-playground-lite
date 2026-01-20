package restart

import (
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/spf13/cobra"
	sysdaemon "github.com/takama/daemon"
)

// Cmd represents the restart command
var Cmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the daemon system service",
	Long:  `Stop and start the P2P Playground daemon system service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := sysdaemon.New(consts.DaemonServiceName, consts.DaemonServiceDescription, sysdaemon.SystemDaemon)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		// Stop first (ignore error if not running)
		if _, err := srv.Stop(); err != nil {
			fmt.Printf("Note: stop returned: %v\n", err)
		}

		// Then start
		status, err := srv.Start()
		if err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println(status)
		return nil
	},
}
