package transfer

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

const (
	protocolID  = "/p2p-playground/transfer/1.0.0"
	chunkSize   = 64 * 1024          // 64KB chunks
	maxFileSize = 1024 * 1024 * 1024 // 1GB max
)

// Manager handles file transfers over P2P
type Manager struct {
	host   types.Host
	logger types.Logger
}

// New creates a new transfer manager
func New(host types.Host, logger types.Logger) *Manager {
	m := &Manager{
		host:   host,
		logger: logger,
	}

	// Set up stream handler for receiving files
	host.SetStreamHandler(protocolID, m.handleIncomingStream)

	return m
}

// Send sends a file to a peer
func (m *Manager) Send(ctx context.Context, peerID string, filePath string, progress types.ProgressCallback) error {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return types.WrapError(err, "failed to open file")
	}
	defer func() { _ = file.Close() }()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return types.WrapError(err, "failed to get file info")
	}

	fileSize := fileInfo.Size()
	if fileSize > maxFileSize {
		return fmt.Errorf("file too large: %d bytes", fileSize)
	}

	// Create stream to peer
	stream, err := m.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return types.WrapError(err, "failed to create stream")
	}
	defer func() { _ = stream.Close() }()

	// Send file size first
	if err := binary.Write(stream, binary.BigEndian, fileSize); err != nil {
		return types.WrapError(err, "failed to send file size")
	}

	// Send file in chunks
	buf := make([]byte, chunkSize)
	var sent int64

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return types.WrapError(err, "failed to read file")
		}

		if n == 0 {
			break
		}

		// Write chunk
		if _, err := stream.Write(buf[:n]); err != nil {
			return types.WrapError(err, "failed to send chunk")
		}

		sent += int64(n)

		// Report progress
		if progress != nil {
			progress(sent, fileSize)
		}
	}

	m.logger.Info("file sent successfully",
		"peer", peerID,
		"file", filePath,
		"size", fileSize,
	)

	return nil
}

// Receive receives a file from a stream
func (m *Manager) Receive(ctx context.Context, stream types.Stream, destPath string, progress types.ProgressCallback) error {
	// Read file size
	var fileSize int64
	if err := binary.Read(stream, binary.BigEndian, &fileSize); err != nil {
		return types.WrapError(err, "failed to read file size")
	}

	if fileSize > maxFileSize {
		return fmt.Errorf("file too large: %d bytes", fileSize)
	}

	// Create destination file
	file, err := os.Create(destPath)
	if err != nil {
		return types.WrapError(err, "failed to create file")
	}
	defer func() { _ = file.Close() }()

	// Receive file in chunks
	buf := make([]byte, chunkSize)
	var received int64

	for received < fileSize {
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return types.WrapError(err, "failed to read chunk")
		}

		if n == 0 {
			break
		}

		// Write to file
		if _, err := file.Write(buf[:n]); err != nil {
			return types.WrapError(err, "failed to write file")
		}

		received += int64(n)

		// Report progress
		if progress != nil {
			progress(received, fileSize)
		}
	}

	if received != fileSize {
		return fmt.Errorf("incomplete transfer: received %d of %d bytes", received, fileSize)
	}

	m.logger.Info("file received successfully",
		"file", destPath,
		"size", fileSize,
	)

	return nil
}

// handleIncomingStream handles incoming transfer streams
func (m *Manager) handleIncomingStream(stream types.Stream) {
	defer func() { _ = stream.Close() }()

	m.logger.Info("incoming file transfer")

	// For now, just close the stream
	// In a real implementation, you'd coordinate with the daemon
	// to determine where to save the file
}
