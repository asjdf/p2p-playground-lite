package start

import (
	"fmt"

	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/spf13/cobra"
	sysdaemon "github.com/takama/daemon"
)

// Cmd represents the start command
var Cmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon system service",
	Long:  `Start the P2P Playground daemon system service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := sysdaemon.New(consts.DaemonServiceName, consts.DaemonServiceDescription, sysdaemon.SystemDaemon)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		status, err := srv.Start()
		if err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println(status)
		return nil
	},
}
