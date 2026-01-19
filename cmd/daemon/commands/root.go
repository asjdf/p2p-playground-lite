package commands

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/asjdf/p2p-playground-lite/pkg/config"
	"github.com/asjdf/p2p-playground-lite/pkg/daemon"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "P2P Playground daemon",
	Long:  `Daemon node for P2P Playground - runs applications distributed from the controller.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Daemon status: Not implemented yet")
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.p2p-playground/daemon.yaml)")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
