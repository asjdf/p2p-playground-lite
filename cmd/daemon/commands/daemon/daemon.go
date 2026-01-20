package daemon

import (
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/install"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/restart"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/run"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/start"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/status"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/stop"
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon/uninstall"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for all daemon-related operations
var Cmd = &cobra.Command{
	Use:   "daemon",
	Short: "Daemon service management",
	Long:  `Manage the P2P Playground daemon service - install, start, stop, and run.`,
}

func init() {
	Cmd.AddCommand(run.Cmd)
	Cmd.AddCommand(install.Cmd)
	Cmd.AddCommand(uninstall.Cmd)
	Cmd.AddCommand(start.Cmd)
	Cmd.AddCommand(stop.Cmd)
	Cmd.AddCommand(restart.Cmd)
	Cmd.AddCommand(status.Cmd)
}
