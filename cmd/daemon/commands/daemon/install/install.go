package install

import (
	"fmt"
	"runtime"

	"github.com/asjdf/p2p-playground-lite/pkg/consts"
	"github.com/spf13/cobra"
	sysdaemon "github.com/takama/daemon"
)

// systemdTemplate is a custom systemd service template with HOME environment variable set
// This is needed because systemd services don't have HOME defined by default
const systemdTemplate = `[Unit]
Description={{.Description}}
Requires={{.Dependencies}}
After={{.Dependencies}}

[Service]
PIDFile=/var/run/{{.Name}}.pid
ExecStartPre=/bin/rm -f /var/run/{{.Name}}.pid
ExecStart={{.Path}} {{.Args}}
Environment="HOME=/root"
WorkingDirectory=/root
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

// Cmd represents the install command
var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install the daemon as a system service",
	Long:  `Install the P2P Playground daemon as a system service (systemd on Linux, launchd on macOS).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := sysdaemon.New(consts.DaemonServiceName, consts.DaemonServiceDescription, sysdaemon.SystemDaemon)
		if err != nil {
			return fmt.Errorf("failed to create daemon: %w", err)
		}

		// Set custom template for Linux systemd to include HOME environment variable
		if runtime.GOOS == "linux" {
			if err := srv.SetTemplate(systemdTemplate); err != nil {
				return fmt.Errorf("failed to set service template: %w", err)
			}
		}

		// Get config file from root command
		cfgFile, _ := cmd.Flags().GetString("config")

		// Build arguments for the service: daemon run + optional config
		// Note: takama/daemon automatically uses the current executable path
		serviceArgs := []string{"daemon", "run"}
		if cfgFile != "" {
			serviceArgs = append(serviceArgs, "-c", cfgFile)
		}

		status, err := srv.Install(serviceArgs...)
		if err != nil {
			return fmt.Errorf("failed to install service: %w", err)
		}

		fmt.Println(status)
		return nil
	},
}
