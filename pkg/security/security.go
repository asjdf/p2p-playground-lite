package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

// Signer implements Ed25519 signing
type Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewSigner creates a new signer
func NewSigner() (*Signer, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, types.WrapError(err, "failed to generate key pair")
	}

	return &Signer{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// LoadSigner loads a signer from key files
func LoadSigner(privKeyPath string) (*Signer, error) {
	privData, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, types.WrapError(err, "failed to read private key")
	}

	if len(privData) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	priv := ed25519.PrivateKey(privData)
	pub := priv.Public().(ed25519.PublicKey)

	return &Signer{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// Sign signs data
func (s *Signer) Sign(data []byte) ([]byte, error) {
	signature := ed25519.Sign(s.privateKey, data)
	return signature, nil
}

// Verify verifies a signature
func (s *Signer) Verify(data []byte, signature []byte, publicKey []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size")
	}

	pub := ed25519.PublicKey(publicKey)
	if !ed25519.Verify(pub, data, signature) {
		return types.ErrInvalidSignature
	}

	return nil
}

// PublicKey returns the public key
func (s *Signer) PublicKey() []byte {
	return s.publicKey
}

// PrivateKey returns the private key (use with caution)
func (s *Signer) PrivateKey() []byte {
	return s.privateKey
}

// SaveKeys saves the key pair to files
func (s *Signer) SaveKeys(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return types.WrapError(err, "failed to create keys directory")
	}

	// Save private key
	privPath := filepath.Join(dir, "node.key")
	if err := os.WriteFile(privPath, s.privateKey, 0600); err != nil {
		return types.WrapError(err, "failed to save private key")
	}

	// Save public key
	pubPath := filepath.Join(dir, "node.pub")
	if err := os.WriteFile(pubPath, s.publicKey, 0644); err != nil {
		return types.WrapError(err, "failed to save public key")
	}

	return nil
}

// GenerateAndSaveKeys generates a new key pair and saves it
func GenerateAndSaveKeys(dir string) (*Signer, error) {
	signer, err := NewSigner()
	if err != nil {
		return nil, err
	}

	if err := signer.SaveKeys(dir); err != nil {
		return nil, err
	}

	return signer, nil
}

// LoadOrGenerateKeys loads keys from directory or generates new ones
func LoadOrGenerateKeys(dir string) (*Signer, error) {
	privPath := filepath.Join(dir, "node.key")

	// Check if keys exist
	if _, err := os.Stat(privPath); err == nil {
		return LoadSigner(privPath)
	}

	// Generate new keys
	return GenerateAndSaveKeys(dir)
}
