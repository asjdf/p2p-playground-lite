package types_test

import (
	"testing"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

func TestAppStatusTypes(t *testing.T) {
	tests := []struct {
		name   string
		status types.AppStatusType
		want   string
	}{
		{"stopped", types.AppStatusStopped, "stopped"},
		{"starting", types.AppStatusStarting, "starting"},
		{"running", types.AppStatusRunning, "running"},
		{"failed", types.AppStatusFailed, "failed"},
		{"restarting", types.AppStatusRestarting, "restarting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

func TestNewAppError(t *testing.T) {
	err := types.NewAppError("TEST_ERROR", "test message", types.ErrNotFound)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "TEST_ERROR" {
		t.Errorf("code = %v, want TEST_ERROR", err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("message = %v, want 'test message'", err.Message)
	}

	if !types.IsNotFoundError(err) {
		t.Error("expected IsNotFoundError to return true")
	}
}

func TestAppErrorWithField(t *testing.T) {
	err := types.NewAppError("TEST_ERROR", "test message", nil)
	err = err.WithField("app_id", "test-app")
	err = err.WithField("count", 42)

	if err.Fields["app_id"] != "test-app" {
		t.Errorf("field app_id = %v, want 'test-app'", err.Fields["app_id"])
	}

	if err.Fields["count"] != 42 {
		t.Errorf("field count = %v, want 42", err.Fields["count"])
	}
}

func TestWrapError(t *testing.T) {
	original := types.ErrNotFound
	wrapped := types.WrapError(original, "failed to find resource")

	if wrapped == nil {
		t.Fatal("expected error, got nil")
	}

	if !types.IsNotFoundError(wrapped) {
		t.Error("expected IsNotFoundError to return true for wrapped error")
	}
}

func TestManifestDefaults(t *testing.T) {
	manifest := &types.Manifest{
		Name:       "test-app",
		Version:    "1.0.0",
		Entrypoint: "bin/app",
	}

	if manifest.Name != "test-app" {
		t.Errorf("name = %v, want 'test-app'", manifest.Name)
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("version = %v, want '1.0.0'", manifest.Version)
	}
}

func TestApplication(t *testing.T) {
	app := &types.Application{
		ID:      "app-123",
		Name:    "test-app",
		Version: "1.0.0",
		Status:  types.AppStatusRunning,
		PID:     12345,
		Labels: map[string]string{
			"env": "test",
		},
	}

	if app.ID != "app-123" {
		t.Errorf("id = %v, want 'app-123'", app.ID)
	}

	if app.Status != types.AppStatusRunning {
		t.Errorf("status = %v, want running", app.Status)
	}

	if app.Labels["env"] != "test" {
		t.Errorf("label env = %v, want 'test'", app.Labels["env"])
	}
}
