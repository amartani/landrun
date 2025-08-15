package elfdeps

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test helpers against a known binary in the system: find `true` via LookPath
func TestParseAndResolveTrue(t *testing.T) {
	bin, err := exec.LookPath("true")
	if err != nil {
		t.Fatalf("failed to find 'true' binary: %v", err)
	}

	f, err := elf.Open(bin)
	if err != nil {
		t.Fatalf("failed to open %s: %v", bin, err)
	}
	defer f.Close()

	interp := parseInterp(f)
	if interp == "" {
		t.Fatalf("expected interpreter for %s, got empty", bin)
	}

	needed, rpaths := parseDynamic(f)
	if needed == nil {
		needed = []string{}
	}

	origin := filepath.Dir(bin)
	paths := resolveSonames(needed, rpaths, origin)
	if paths == nil {
		paths = []string{}
	}

	// Ensure interpreter path exists on filesystem
	if _, err := os.Stat(interp); err != nil {
		t.Fatalf("interp path %s does not exist: %v", interp, err)
	}

	// If there are resolved library paths, they must exist
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("resolved library path %s does not exist: %v", p, err)
		}
	}
}

func TestGetLibraryDependencies(t *testing.T) {
	bin, err := exec.LookPath("true")
	if err != nil {
		t.Fatalf("failed to find 'true' binary: %v", err)
	}
	paths, err := GetLibraryDependencies(bin)
	if err != nil {
		t.Fatalf("GetLibraryDependencies failed: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("expected non-empty dependency list for %s", bin)
	}
	// ensure returned paths are absolute and exist
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			t.Fatalf("expected absolute path, got %s", p)
		}
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("path %s does not exist: %v", p, err)
		}
	}
}

func TestGetLdmapWithFakeOutput(t *testing.T) {
	// fake ldconfig output with a single mapping
	original := ldconfigRunner
	defer func() { ldconfigRunner = original }()

	// create a fake file on disk to satisfy os.Stat checks in getLdmap
	tmpDir := t.TempDir()
	tmp := filepath.Join(tmpDir, "libfake.so")
	f, err := os.Create(tmp)
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	f.Close()

	// Because getLdmap checks the path exists, return tmp in the fake output
	ldconfigRunner = func() ([]byte, error) {
		return []byte("libfake.so (libc6,x86-64) => " + tmp + "\n"), nil
	}

	m := getLdmap()
	if got, ok := m["libfake.so"]; !ok {
		t.Fatalf("expected libfake.so in map")
	} else if got != tmp {
		t.Fatalf("expected path %s, got %s", tmp, got)
	}
}

func TestResolveSonamesUsesLdmapFallback(t *testing.T) {
	original := ldconfigRunner
	defer func() { ldconfigRunner = original }()

	tmpDir := t.TempDir()
	tmp := filepath.Join(tmpDir, "libfake2.so")
	f, err := os.Create(tmp)
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	f.Close()

	ldconfigRunner = func() ([]byte, error) {
		return []byte("libfake2.so (libc6,x86-64) => " + tmp + "\n"), nil
	}

	// needed contains a soname that won't be found in rpaths or std dirs
	out := resolveSonames([]string{"libfake2.so"}, []string{}, tmpDir)
	if len(out) != 1 {
		t.Fatalf("expected 1 resolved path, got %d", len(out))
	}
	if out[0] != tmp {
		t.Fatalf("expected %s, got %s", tmp, out[0])
	}
}

func TestResolveSonamesOriginExpansion(t *testing.T) {
	// Create a temp dir and a lib subdir to simulate $ORIGIN/lib
	tmpDir := t.TempDir()
	libDir := filepath.Join(tmpDir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatalf("failed create lib dir: %v", err)
	}

	libName := "liborigin.so"
	libPath := filepath.Join(libDir, libName)
	f, err := os.Create(libPath)
	if err != nil {
		t.Fatalf("failed to create lib file: %v", err)
	}
	f.Close()

	// rpath using $ORIGIN should resolve to tmpDir/lib
	out := resolveSonames([]string{libName}, []string{"$ORIGIN/lib"}, tmpDir)
	if len(out) != 1 {
		t.Fatalf("expected 1 resolved path for $ORIGIN, got %d", len(out))
	}
	if out[0] != libPath {
		t.Fatalf("expected %s, got %s", libPath, out[0])
	}

	// relative rpath should also resolve against origin
	out2 := resolveSonames([]string{libName}, []string{"lib"}, tmpDir)
	if len(out2) != 1 {
		t.Fatalf("expected 1 resolved path for relative rpath, got %d", len(out2))
	}
	if out2[0] != libPath {
		t.Fatalf("expected %s, got %s", libPath, out2[0])
	}
}
