package commands

import (
	"github.com/asjdf/p2p-playground-lite/cmd/daemon/commands/daemon"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "p2p-daemon",
	Short: "P2P Playground daemon CLI",
	Long:  `P2P Playground daemon CLI - manage the P2P Playground daemon service.`,
}

// GetCfgFile returns the config file path
func GetCfgFile() string {
	return cfgFile
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.p2p-playground/daemon.yaml)")

	// Add daemon command
	rootCmd.AddCommand(daemon.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
