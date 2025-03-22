package sandbox

import (
	"fmt"

	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/zouuup/landrun/internal/log"
)

type Config struct {
	ReadOnlyPaths   []string
	ReadWritePaths  []string
	AllowExec       bool
	BindTCPPorts    []int
	ConnectTCPPorts []int
	BestEffort      bool
}

func Apply(cfg Config) error {
	log.Info("Sandbox config: %+v", cfg)

	// Get the most advanced Landlock version available
	llCfg := landlock.V5
	if cfg.BestEffort {
		llCfg = llCfg.BestEffort()
	}

	// Collect our rules
	var rules []landlock.Rule

	// Add rules for read-only paths
	for _, path := range cfg.ReadOnlyPaths {
		log.Debug("Adding read-only path: %s", path)
		if cfg.AllowExec {
			// Include execution rights for read-only paths
			rules = append(rules, landlock.RODirs(path))
		} else {
			// Use ROFiles which doesn't include execution rights
			rules = append(rules, landlock.ROFiles(path))
		}
	}

	// Add rules for read-write paths
	for _, path := range cfg.ReadWritePaths {
		log.Debug("Adding read-write path: %s", path)
		if cfg.AllowExec {
			// Include all permissions for read-write paths
			rules = append(rules, landlock.RWDirs(path))
		} else {
			// Use RWFiles which doesn't include execution rights
			rules = append(rules, landlock.RWFiles(path))
		}
	}

	// Add rules for TCP port binding
	for _, port := range cfg.BindTCPPorts {
		log.Debug("Adding TCP bind port: %d", port)
		rules = append(rules, landlock.BindTCP(uint16(port)))
	}

	// Add rules for TCP connections
	for _, port := range cfg.ConnectTCPPorts {
		log.Debug("Adding TCP connect port: %d", port)
		rules = append(rules, landlock.ConnectTCP(uint16(port)))
	}

	// If we have no rules, just return
	if len(rules) == 0 {
		log.Error("No rules provided, applying default restrictive rules, this will restrict anything landlock can do.")
		err := llCfg.Restrict()
		if err != nil {
			return fmt.Errorf("failed to apply default Landlock restrictions: %w", err)
		}
		log.Info("Default restrictive Landlock rules applied successfully")
		return nil
	}

	// Apply all rules at once
	log.Debug("Applying Landlock restrictions")
	err := llCfg.Restrict(rules...)
	if err != nil {
		return fmt.Errorf("failed to apply Landlock restrictions: %w", err)
	}

	log.Info("Landlock restrictions applied successfully")
	return nil
}
