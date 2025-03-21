package sandbox

import (
	"fmt"
	"github.com/armin/landrun/internal/landlock"
)

type Config struct {
	ReadOnlyPaths  []string
	ReadWritePaths []string
	AllowExec      bool
}

func Apply(cfg Config) error {
	// TODO: use landlock.CreateRuleset, AddRule, RestrictSelf
	fmt.Println("[landrun] Sandbox config:", cfg)

	// Example: fail if no kernel support (stubbed)
	if !landlock.IsSupported() {
		return fmt.Errorf("Landlock not supported on this system")
	}

	// TODO: Build ruleset from cfg.ReadOnlyPaths, cfg.ReadWritePaths
	// landlock.CreateRuleset(...)
	// landlock.AddRule(...)
	// landlock.RestrictSelf(...)

	return nil
}
