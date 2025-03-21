package sandbox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApply(t *testing.T) {
	// Create test directories in current directory
	roDir := "test_ro"
	rwDir := "test_rw"

	// Clean up any existing test directories
	os.RemoveAll(roDir)
	os.RemoveAll(rwDir)

	// Create test directories
	if err := os.MkdirAll(roDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rwDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	if err := os.WriteFile(filepath.Join(roDir, "test.txt"), []byte("readonly"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rwDir, "test.txt"), []byte("readwrite"), 0644); err != nil {
		t.Fatal(err)
	}

	// Clean up after tests
	defer func() {
		os.RemoveAll(roDir)
		os.RemoveAll(rwDir)
	}()

	tests := []struct {
		name        string
		cfg         Config
		expectError bool
	}{
		{
			name: "valid config with read-only and read-write paths",
			cfg: Config{
				ReadOnlyPaths:  []string{roDir},
				ReadWritePaths: []string{rwDir},
				AllowExec:      false,
			},
			expectError: false,
		},
		{
			name: "valid config with exec allowed",
			cfg: Config{
				ReadOnlyPaths:  []string{roDir},
				ReadWritePaths: []string{rwDir},
				AllowExec:      true,
			},
			expectError: false,
		},
		{
			name: "empty config",
			cfg: Config{
				ReadOnlyPaths:  []string{},
				ReadWritePaths: []string{},
				AllowExec:      false,
			},
			expectError: false,
		},
		{
			name: "non-existent read-only path",
			cfg: Config{
				ReadOnlyPaths:  []string{"/nonexistent/path"},
				ReadWritePaths: []string{rwDir},
				AllowExec:      false,
			},
			expectError: true,
		},
		{
			name: "non-existent read-write path",
			cfg: Config{
				ReadOnlyPaths:  []string{roDir},
				ReadWritePaths: []string{"/nonexistent/path"},
				AllowExec:      false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Apply(tt.cfg)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
