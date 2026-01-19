package types

import (
	"errors"
	"fmt"
)

// Common errors that can be returned by various components
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput indicates invalid input was provided
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates the operation is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("timeout")

	// ErrCanceled indicates an operation was canceled
	ErrCanceled = errors.New("canceled")

	// ErrInternal indicates an internal error occurred
	ErrInternal = errors.New("internal error")

	// ErrNotImplemented indicates functionality is not implemented
	ErrNotImplemented = errors.New("not implemented")

	// ErrUnavailable indicates a resource is temporarily unavailable
	ErrUnavailable = errors.New("unavailable")

	// ErrInvalidState indicates an operation cannot be performed in the current state
	ErrInvalidState = errors.New("invalid state")
)

// Application-specific errors
var (
	// ErrAppNotRunning indicates an application is not running
	ErrAppNotRunning = errors.New("application not running")

	// ErrAppAlreadyRunning indicates an application is already running
	ErrAppAlreadyRunning = errors.New("application already running")

	// ErrAppStartFailed indicates an application failed to start
	ErrAppStartFailed = errors.New("application start failed")

	// ErrAppStopFailed indicates an application failed to stop
	ErrAppStopFailed = errors.New("application stop failed")

	// ErrAppUnhealthy indicates an application failed health checks
	ErrAppUnhealthy = errors.New("application unhealthy")
)

// Package-specific errors
var (
	// ErrInvalidManifest indicates a manifest file is invalid
	ErrInvalidManifest = errors.New("invalid manifest")

	// ErrInvalidPackage indicates a package file is invalid or corrupted
	ErrInvalidPackage = errors.New("invalid package")

	// ErrInvalidSignature indicates a signature verification failed
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidChecksum indicates a checksum verification failed
	ErrInvalidChecksum = errors.New("invalid checksum")

	// ErrPackageNotSigned indicates a package is not signed
	ErrPackageNotSigned = errors.New("package not signed")
)

// P2P-specific errors
var (
	// ErrPeerNotFound indicates a peer was not found
	ErrPeerNotFound = errors.New("peer not found")

	// ErrConnectionFailed indicates a connection attempt failed
	ErrConnectionFailed = errors.New("connection failed")

	// ErrStreamClosed indicates a stream was closed
	ErrStreamClosed = errors.New("stream closed")

	// ErrProtocolNotSupported indicates a protocol is not supported
	ErrProtocolNotSupported = errors.New("protocol not supported")
)

// Version-specific errors
var (
	// ErrInvalidVersion indicates a version string is invalid
	ErrInvalidVersion = errors.New("invalid version")

	// ErrVersionConflict indicates a version conflict
	ErrVersionConflict = errors.New("version conflict")

	// ErrNoVersionsAvailable indicates no versions are available
	ErrNoVersionsAvailable = errors.New("no versions available")
)

// Storage-specific errors
var (
	// ErrStorageRead indicates a storage read operation failed
	ErrStorageRead = errors.New("storage read failed")

	// ErrStorageWrite indicates a storage write operation failed
	ErrStorageWrite = errors.New("storage write failed")

	// ErrStorageDelete indicates a storage delete operation failed
	ErrStorageDelete = errors.New("storage delete failed")
)

// AppError provides structured error information
type AppError struct {
	// Code is the error code
	Code string

	// Message is the error message
	Message string

	// Err is the underlying error
	Err error

	// Fields contains additional context
	Fields map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewAppError creates a new AppError
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// WithField adds a field to the error
func (e *AppError) WithField(key string, value interface{}) *AppError {
	e.Fields[key] = value
	return e
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExistsError checks if an error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInputError checks if an error is an invalid input error
func IsInvalidInputError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}
