package sandbox

import (
	"fmt"

	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/landlock-lsm/go-landlock/landlock/syscall"
	"github.com/zouuup/landrun/internal/log"
)

type Config struct {
	ReadOnlyPaths   []string
	ReadWritePaths  []string
	ExecutablePaths []string
	BindTCPPorts    []int
	ConnectTCPPorts []int
	BestEffort      bool
}

// getExecutableRights returns a full set of permissions including execution
func getExecutableRights() landlock.AccessFSSet {
	accessRights := landlock.AccessFSSet(0)
	// Add execute permission
	accessRights |= landlock.AccessFSSet(syscall.AccessFSExecute)
	// Add read permissions
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSReadDir)
	// Add write permissions
	accessRights |= landlock.AccessFSSet(syscall.AccessFSWriteFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSRemoveFile)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeChar)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeDir)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeReg)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeFifo)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeBlock)
	accessRights |= landlock.AccessFSSet(syscall.AccessFSMakeSym)
	return accessRights
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

	// Process executable paths first - these need special handling
	for _, path := range cfg.ExecutablePaths {
		log.Debug("Adding executable path: %s", path)
		// Use PathAccess with all permissions including execute
		rules = append(rules, landlock.PathAccess(getExecutableRights(), path))
	}

	// Process read-only paths
	for _, path := range cfg.ReadOnlyPaths {
		// Skip if already handled as executable
		if pathInSlice(path, cfg.ExecutablePaths) {
			log.Debug("Skipping read-only path (already executable): %s", path)
			continue
		}

		log.Debug("Adding read-only path: %s", path)
		// Use RODirs which includes the ability to read files and directories
		rules = append(rules, landlock.RODirs(path))
	}

	// Process read-write paths
	for _, path := range cfg.ReadWritePaths {
		// Skip if already handled as executable
		if pathInSlice(path, cfg.ExecutablePaths) {
			log.Debug("Skipping read-write path (already executable): %s", path)
			continue
		}

		log.Debug("Adding read-write path: %s", path)
		// Use RWDirs which includes full read/write permissions
		rules = append(rules, landlock.RWDirs(path))
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

// pathInSlice checks if a path exists in a slice of paths
func pathInSlice(path string, paths []string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
