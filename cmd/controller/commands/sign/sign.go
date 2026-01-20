package sign

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/security"
	"github.com/spf13/cobra"
)

var (
	keyPath string
)

// Cmd represents the sign command
var Cmd = &cobra.Command{
	Use:   "sign [package]",
	Short: "Sign an application package",
	Long: `Sign an application package with your private key.

The signature will be embedded in the deployment request and verified by nodes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packagePath := args[0]

		// Check if package exists
		if _, err := os.Stat(packagePath); err != nil {
			return fmt.Errorf("package not found: %w", err)
		}

		// Determine key path
		kp := keyPath
		if kp == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			kp = filepath.Join(home, ".p2p-playground", "keys", "controller.key")
		}

		// Load private key
		signer, err := security.LoadSigner(kp)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}

		// Sign package
		fmt.Printf("Signing package: %s\n", packagePath)
		signature, err := signer.SignFile(packagePath)
		if err != nil {
			return fmt.Errorf("failed to sign package: %w", err)
		}

		// Save signature to file
		sigPath := packagePath + ".sig"
		if err := os.WriteFile(sigPath, signature, 0644); err != nil {
			return fmt.Errorf("failed to save signature: %w", err)
		}

		fmt.Printf("\n✓ Package signed successfully!\n")
		fmt.Printf("  Signature: %s\n", sigPath)
		fmt.Printf("  Signature (hex): %s\n", hex.EncodeToString(signature))
		fmt.Printf("\n")
		fmt.Printf("You can now deploy this package with signature verification.\n")

		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&keyPath, "key", "k", "", "path to private key file (default: ~/.p2p-playground/keys/controller.key)")
}
