package logger

import (
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "development environment returns slog.Logger",
			env:  "development",
		},
		{
			name: "production environment returns slog.Logger",
			env:  "production",
		},
		{
			name: "unknown environment defaults to development behavior",
			env:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.env)
			if l == nil {
				t.Fatal("expected non-nil *slog.Logger, got nil")
			}
			var _ *slog.Logger = l // compile-time type assertion
		})
	}
}
