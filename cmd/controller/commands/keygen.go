package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/spf13/cobra"
)

var (
	keygenOutput string
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate Ed25519 key pair for package signing",
	Long: `Generate a new Ed25519 key pair for signing application packages.

The private key (controller.key) is used to sign packages.
The public key (controller.pub) should be distributed to nodes for signature verification.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine output directory
		outputDir := keygenOutput
		if outputDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			outputDir = filepath.Join(home, ".p2p-playground", "keys")
		}

		// Generate keys
		fmt.Printf("Generating Ed25519 key pair...\n")
		signer, err := security.GenerateAndSaveKeys(outputDir, "controller")
		if err != nil {
			return fmt.Errorf("failed to generate keys: %w", err)
		}

		fmt.Printf("\n✓ Key pair generated successfully!\n")
		fmt.Printf("  Private key: %s/controller.key\n", outputDir)
		fmt.Printf("  Public key:  %s/controller.pub\n", outputDir)
		fmt.Printf("\n")
		fmt.Printf("⚠️  Keep the private key secure and never share it.\n")
		fmt.Printf("📤 Distribute the public key to nodes for signature verification.\n")
		fmt.Printf("\n")
		fmt.Printf("Public key (hex): %x\n", signer.PublicKey())

		return nil
	},
}

func init() {
	keygenCmd.Flags().StringVarP(&keygenOutput, "output", "o", "", "output directory for keys (default: ~/.p2p-playground/keys)")
}
