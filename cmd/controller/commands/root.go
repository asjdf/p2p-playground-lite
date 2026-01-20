package commands

import (
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/common"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/deploy"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/keygen"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/list"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/logs"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/nodes"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/psk"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/run"
	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands/sign"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "controller",
	Short: "P2P Playground controller",
	Long:  `Controller for P2P Playground - deploy and manage applications across P2P nodes.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return common.InitConfig(cfgFile)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.p2p-playground/controller.yaml)")

	rootCmd.AddCommand(deploy.Cmd)
	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(logs.Cmd)
	rootCmd.AddCommand(nodes.Cmd)
	rootCmd.AddCommand(run.Cmd)
	rootCmd.AddCommand(keygen.Cmd)
	rootCmd.AddCommand(sign.Cmd)
	rootCmd.AddCommand(psk.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
