package run

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/daemon"
	"github.com/spf13/cobra"
)

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon in foreground",
	Long:  `Run the P2P Playground daemon in foreground mode. Use Ctrl+C to stop.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config file from root command
		cfgFile, _ := cmd.Flags().GetString("config")

		// Load config
		cfg, err := config.LoadDaemonConfig(cfgFile)
		if err != nil {
			return err
		}

		// Create daemon
		d, err := daemon.New(cfg)
		if err != nil {
			return err
		}

		// Start daemon
		if err := d.Start(); err != nil {
			return err
		}

		// Wait for signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan

		// Stop daemon
		return d.Stop()
	},
}
