package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

const (
	// PSKSize is the size of a PSK in bytes (32 bytes = 256 bits)
	PSKSize = 32
)

// GeneratePSK generates a new pre-shared key
func GeneratePSK() ([]byte, error) {
	psk := make([]byte, PSKSize)
	if _, err := rand.Read(psk); err != nil {
		return nil, types.WrapError(err, "failed to generate PSK")
	}
	return psk, nil
}

// EncodePSK encodes a PSK as a hex string
func EncodePSK(psk []byte) string {
	return hex.EncodeToString(psk)
}

// DecodePSK decodes a hex-encoded PSK
func DecodePSK(encoded string) ([]byte, error) {
	psk, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, types.WrapError(err, "failed to decode PSK")
	}

	if len(psk) != PSKSize {
		return nil, fmt.Errorf("invalid PSK size: expected %d bytes, got %d", PSKSize, len(psk))
	}

	return psk, nil
}

// SavePSK saves a PSK to a file
func SavePSK(psk []byte, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return types.WrapError(err, "failed to create directory")
	}

	// Encode PSK as hex and save to file
	encoded := EncodePSK(psk)
	if err := os.WriteFile(filePath, []byte(encoded), 0600); err != nil {
		return types.WrapError(err, "failed to write PSK file")
	}

	return nil
}

// LoadPSK loads a PSK from a file
func LoadPSK(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, types.WrapError(err, "failed to read PSK file")
	}

	return DecodePSK(string(data))
}

// GenerateAndSavePSK generates a new PSK and saves it to a file
func GenerateAndSavePSK(filePath string) ([]byte, error) {
	psk, err := GeneratePSK()
	if err != nil {
		return nil, err
	}

	if err := SavePSK(psk, filePath); err != nil {
		return nil, err
	}

	return psk, nil
}
