package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/spf13/cobra"
)

var (
	pskOutputPath string
)

var pskCmd = &cobra.Command{
	Use:   "psk",
	Short: "Generate a pre-shared key for private P2P network",
	Long: `Generate a pre-shared key (PSK) for creating a private P2P network.

The PSK can be used to ensure that only authorized nodes can join your network.
All nodes (controller and daemons) must use the same PSK to communicate.

Example:
  controller psk
  controller psk --output ~/.p2p-playground/psk`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Generating pre-shared key...")

		// Generate PSK
		psk, err := security.GeneratePSK()
		if err != nil {
			return fmt.Errorf("failed to generate PSK: %w", err)
		}

		// Determine output path
		outputPath := pskOutputPath
		if outputPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			outputPath = filepath.Join(homeDir, ".p2p-playground", "psk")
		}

		// Save PSK to file
		if err := security.SavePSK(psk, outputPath); err != nil {
			return fmt.Errorf("failed to save PSK: %w", err)
		}

		encoded := security.EncodePSK(psk)

		fmt.Println()
		fmt.Println("✓ PSK generated successfully!")
		fmt.Printf("  File: %s\n", outputPath)
		fmt.Println()
		fmt.Println("⚠️  Keep this key secure and never share it publicly.")
		fmt.Println("📤 Distribute this key to all nodes that should join your network.")
		fmt.Println()
		fmt.Printf("PSK (hex): %s\n", encoded)
		fmt.Println()
		fmt.Println("To use this PSK, add the following to your configuration:")
		fmt.Println()
		fmt.Println("  security:")
		fmt.Println("    enable_auth: true")
		fmt.Printf("    psk: \"%s\"\n", encoded)
		fmt.Println()

		return nil
	},
}

func init() {
	pskCmd.Flags().StringVarP(&pskOutputPath, "output", "o", "", "output path for PSK file (default: ~/.p2p-playground/psk)")
	rootCmd.AddCommand(pskCmd)
}
