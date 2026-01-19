package pkgmanager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
	"gopkg.in/yaml.v3"
)

// Manager implements package management
type Manager struct{}

// New creates a new package manager
func New() *Manager {
	return &Manager{}
}

// Pack creates a tar.gz package from an application directory
func (m *Manager) Pack(ctx context.Context, appDir string) (string, error) {
	// Read manifest
	manifest, err := m.readManifest(filepath.Join(appDir, "manifest.yaml"))
	if err != nil {
		return "", err
	}

	// Create output package path
	pkgName := fmt.Sprintf("%s-%s.tar.gz", manifest.Name, manifest.Version)
	pkgPath := filepath.Join(filepath.Dir(appDir), pkgName)

	// Create tar.gz file
	outFile, err := os.Create(pkgPath)
	if err != nil {
		return "", types.WrapError(err, "failed to create package file")
	}
	defer func() { _ = outFile.Close() }()

	gzWriter := gzip.NewWriter(outFile)
	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	// Walk directory and add files
	err = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(appDir, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if not a directory
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", types.WrapError(err, "failed to pack directory")
	}

	return pkgPath, nil
}

// Unpack extracts a package to a destination directory
func (m *Manager) Unpack(ctx context.Context, pkgPath string, destDir string) (*types.Manifest, error) {
	// Open package file
	file, err := os.Open(pkgPath)
	if err != nil {
		return nil, types.WrapError(err, "failed to open package")
	}
	defer func() { _ = file.Close() }()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, types.WrapError(err, "invalid gzip format")
	}
	defer func() { _ = gzReader.Close() }()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	var manifest *types.Manifest
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, types.WrapError(err, "failed to read tar")
		}

		// Create target path
		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, types.WrapError(err, "failed to create directory")
			}

		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, types.WrapError(err, "failed to create parent dir")
			}

			// Create file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return nil, types.WrapError(err, "failed to create file")
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return nil, types.WrapError(err, "failed to write file")
			}
			_ = outFile.Close()

			// Read manifest if this is the manifest file
			if header.Name == "manifest.yaml" {
				manifest, _ = m.readManifest(target)
			}
		}
	}

	if manifest == nil {
		return nil, types.ErrInvalidManifest
	}

	return manifest, nil
}

// Verify verifies a package's integrity
func (m *Manager) Verify(ctx context.Context, pkgPath string, signature []byte) error {
	// For now, just check if file exists and is readable
	// TODO: Implement signature verification
	_, err := os.Stat(pkgPath)
	if err != nil {
		return types.WrapError(err, "package not found")
	}
	return nil
}

// GetManifest reads and parses a manifest from a package
func (m *Manager) GetManifest(ctx context.Context, pkgPath string) (*types.Manifest, error) {
	// Open package
	file, err := os.Open(pkgPath)
	if err != nil {
		return nil, types.WrapError(err, "failed to open package")
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, types.WrapError(err, "invalid gzip format")
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)

	// Find and read manifest.yaml
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, types.WrapError(err, "failed to read tar")
		}

		if header.Name == "manifest.yaml" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, types.WrapError(err, "failed to read manifest")
			}

			var manifest types.Manifest
			if err := yaml.Unmarshal(data, &manifest); err != nil {
				return nil, types.WrapError(err, "failed to parse manifest")
			}

			return &manifest, nil
		}
	}

	return nil, types.ErrInvalidManifest
}

// readManifest reads a manifest file
func (m *Manager) readManifest(path string) (*types.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, types.WrapError(err, "failed to read manifest")
	}

	var manifest types.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, types.WrapError(err, "failed to parse manifest")
	}

	// Validate required fields
	if manifest.Name == "" {
		return nil, fmt.Errorf("manifest missing name: %w", types.ErrInvalidManifest)
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf("manifest missing version: %w", types.ErrInvalidManifest)
	}
	if manifest.Entrypoint == "" {
		return nil, fmt.Errorf("manifest missing entrypoint: %w", types.ErrInvalidManifest)
	}

	return &manifest, nil
}

// CalculateChecksum calculates SHA-256 checksum of a package
func (m *Manager) CalculateChecksum(pkgPath string) (string, error) {
	file, err := os.Open(pkgPath)
	if err != nil {
		return "", types.WrapError(err, "failed to open package")
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", types.WrapError(err, "failed to calculate checksum")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
