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
	if !landlock.IsSupported() {
		log.Fatal("Landlock is not supported or enabled on this system")
	}

	log.Info("Sandbox config: %+v", cfg)

	accessMask := uint64(landlock.AccessReadFile | landlock.AccessReadDir)
	if cfg.AllowExec {
		accessMask |= landlock.AccessExecute
	}

	// Write permissions
	rwAccess := accessMask | landlock.AccessWriteFile |
		landlock.AccessRemoveDir | landlock.AccessRemoveFile |
		landlock.AccessMakeChar | landlock.AccessMakeDir |
		landlock.AccessMakeReg | landlock.AccessMakeSock |
		landlock.AccessMakeFifo | landlock.AccessMakeBlock |
		landlock.AccessMakeSym

	rulesetFd, err := landlock.CreateRuleset(accessMask | rwAccess)
	if err != nil {
		return fmt.Errorf("failed to create Landlock ruleset: %w", err)
	}
	defer landlock.CloseFd(rulesetFd)

	for _, path := range cfg.ReadOnlyPaths {
		log.Debug("Adding read-only path: %s", path)
		err := landlock.AddPathRule(rulesetFd, path, accessMask)
		if err != nil {
			return fmt.Errorf("failed to add read-only rule for %s: %w", path, err)
		}
	}

	for _, path := range cfg.ReadWritePaths {
		log.Debug("Adding read-write path: %s", path)
		err := landlock.AddPathRule(rulesetFd, path, rwAccess)
		if err != nil {
			return fmt.Errorf("failed to add read-write rule for %s: %w", path, err)
		}
	}

	if err := landlock.RestrictSelf(rulesetFd); err != nil {
		return fmt.Errorf("failed to restrict self: %w", err)
	}

	log.Info("Landlock ruleset applied successfully")
	return nil
}
