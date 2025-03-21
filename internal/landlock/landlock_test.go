package landlock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateRuleset(t *testing.T) {
	tests := []struct {
		name        string
		accessMask  uint64
		expectError bool
	}{
		{
			name:        "valid access mask",
			accessMask:  AccessReadFile | AccessReadDir,
			expectError: false,
		},
		{
			name:        "zero access mask",
			accessMask:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, err := CreateRuleset(tt.accessMask)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fd < 0 {
				t.Error("expected valid file descriptor")
			}
			CloseFd(fd)
		})
	}
}

func TestAddPathRule(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		path        string
		accessMask  uint64
		expectError bool
	}{
		{
			name:        "valid path",
			path:        testFile,
			accessMask:  AccessReadFile,
			expectError: false,
		},
		{
			name:        "non-existent path",
			path:        "/nonexistent/path",
			accessMask:  AccessReadFile,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesetFd, err := CreateRuleset(AccessReadFile)
			if err != nil {
				t.Fatal(err)
			}
			defer CloseFd(rulesetFd)

			err = AddPathRule(rulesetFd, tt.path, tt.accessMask)
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

func TestRestrictSelf(t *testing.T) {
	tests := []struct {
		name        string
		accessMask  uint64
		expectError bool
	}{
		{
			name:        "valid ruleset",
			accessMask:  AccessReadFile | AccessReadDir,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesetFd, err := CreateRuleset(tt.accessMask)
			if err != nil {
				t.Fatal(err)
			}
			defer CloseFd(rulesetFd)

			err = RestrictSelf(rulesetFd)
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

func TestCloseFd(t *testing.T) {
	tests := []struct {
		name        string
		fd          int
		expectError bool
	}{
		{
			name:        "invalid file descriptor",
			fd:          -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CloseFd(tt.fd)
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
