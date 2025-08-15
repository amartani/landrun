package elfdeps

import (
	"debug/elf"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
)

// ldconfigRunner runs `ldconfig -p` and returns its output. Tests may override
// this variable to inject fake output. It is unexported on purpose to allow
// test injection within the package.
var ldconfigRunner = func() ([]byte, error) {
	return osexec.Command("ldconfig", "-p").Output()
}

// getLdmap runs `ldconfig -p` and returns a map of soname -> path.
func getLdmap() map[string]string {
	m := map[string]string{}
	out, err := ldconfigRunner()
	if err != nil {
		return m
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "=>") {
			continue
		}
		parts := strings.Split(line, "=>")
		if len(parts) < 2 {
			continue
		}
		path := strings.TrimSpace(parts[len(parts)-1])
		left := strings.TrimSpace(parts[0])
		toks := strings.Fields(left)
		if len(toks) == 0 {
			continue
		}
		soname := toks[0]
		if path == "" || soname == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			if _, exists := m[soname]; !exists {
				m[soname] = path
			}
		}
	}
	return m
}

// parseInterp extracts the PT_INTERP interpreter path from an ELF file.
func parseInterp(f *elf.File) string {
	interp := ""
	for _, prog := range f.Progs {
		if prog.Type == elf.PT_INTERP {
			r := prog.Open()
			if r == nil {
				// Can't read interpreter; return what we have (empty)
				break
			}
			if data, err := io.ReadAll(r); err == nil {
				interp = strings.TrimRight(string(data), "\x00")
			}
			break
		}
	}
	return interp
}

// parseDynamic extracts DT_NEEDED and RPATH/RUNPATH entries from the .dynamic section.
func parseDynamic(f *elf.File) (needed []string, rpaths []string) {
	needed = []string{}
	rpaths = []string{}

	if libs, err := f.DynString(elf.DT_NEEDED); err == nil {
		needed = append(needed, libs...)
	}

	// DT_RPATH and DT_RUNPATH may both be present; split on ':' and append
	if rp, err := f.DynString(elf.DT_RPATH); err == nil {
		for _, v := range rp {
			if v == "" {
				continue
			}
			rpaths = append(rpaths, strings.Split(v, ":")...)
		}
	}
	if rp, err := f.DynString(elf.DT_RUNPATH); err == nil {
		for _, v := range rp {
			if v == "" {
				continue
			}
			rpaths = append(rpaths, strings.Split(v, ":")...)
		}
	}
	return
}

// resolveSonames attempts to resolve sonames to absolute paths using rpaths,
// standard library directories and falling back to parsing `ldconfig -p` output.
// The origin parameter should be the directory containing the binary (used to
// expand $ORIGIN and resolve relative RPATH entries).
func resolveSonames(needed []string, rpaths []string, origin string) []string {
	resolved := map[string]string{}
	seen := map[string]struct{}{}

	stdDirs := []string{"/lib", "/lib64", "/usr/lib", "/usr/lib64", "/usr/local/lib"}

	var ldmap map[string]string

	var resolveOne func(string) string
	resolveOne = func(soname string) string {
		if p, ok := resolved[soname]; ok {
			return p
		}
		if _, s := seen[soname]; s {
			return ""
		}
		seen[soname] = struct{}{}

		candidates := []string{}
		for _, rp := range rpaths {
			if rp == "" {
				continue
			}
			// expand $ORIGIN (common token in RPATH/RUNPATH)
			rp = strings.ReplaceAll(rp, "$ORIGIN", origin)
			rp = strings.ReplaceAll(rp, "${ORIGIN}", origin)
			// make relative rpath entries absolute using origin
			if !filepath.IsAbs(rp) {
				rp = filepath.Join(origin, rp)
			}
			candidates = append(candidates, filepath.Join(rp, soname))
		}
		for _, d := range stdDirs {
			candidates = append(candidates, filepath.Join(d, soname))
		}

		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				resolved[soname] = cand
				return cand
			}
		}

		// fallback: consult parsed ldconfig map (populate lazily)
		if ldmap == nil {
			ldmap = getLdmap()
		}
		if p, ok := ldmap[soname]; ok {
			resolved[soname] = p
			return p
		}

		return ""
	}

	out := []string{}
	for _, s := range needed {
		if p := resolveOne(s); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// GetLibraryDependencies returns a list of library paths that the given binary depends on
func GetLibraryDependencies(binary string) ([]string, error) {
	f, err := elf.Open(binary)
	if err != nil {
		return nil, fmt.Errorf("open ELF %s: %w", binary, err)
	}
	defer f.Close()

	interpPath := parseInterp(f)
	needed, rpaths := parseDynamic(f)
	origin := filepath.Dir(binary)
	libPaths := resolveSonames(needed, rpaths, origin)

	finalPaths := []string{}
	if interpPath != "" {
		finalPaths = append(finalPaths, interpPath)
	}
	finalPaths = append(finalPaths, libPaths...)

	// Add /etc/ld.so.cache if present
	if _, err := os.Stat("/etc/ld.so.cache"); err == nil {
		finalPaths = append(finalPaths, "/etc/ld.so.cache")
	}

	// Flatten unique list
	out := []string{}
	seenOut := map[string]struct{}{}
	for _, p := range finalPaths {
		if _, ok := seenOut[p]; !ok {
			out = append(out, p)
			seenOut[p] = struct{}{}
		}
	}

	return out, nil
}
