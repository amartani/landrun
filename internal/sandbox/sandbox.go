package sandbox

import (
	"fmt"

	"github.com/zouuup/landrun/internal/landlock"
	"github.com/zouuup/landrun/internal/log"
)

type Config struct {
	ReadOnlyPaths  []string
	ReadWritePaths []string
	AllowExec      bool
}

func Apply(cfg Config) error {
	log.Info("Sandbox config: %+v", cfg)

	if !landlock.IsSupported() {
		return fmt.Errorf("landlock not supported on this system")
	}

	// TODO: Build ruleset from cfg.ReadOnlyPaths, cfg.ReadWritePaths
	// landlock.CreateRuleset(...)
	// landlock.AddRule(...)
	// landlock.RestrictSelf(...)

	return nil
}
