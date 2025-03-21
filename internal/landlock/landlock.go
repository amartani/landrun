package landlock

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ---------- Constants from <linux/landlock.h> ----------

const (
	// Access rights (bitfields)
	AccessExecute           = 1 << 0
	AccessWriteFile         = 1 << 1
	AccessReadFile          = 1 << 2
	AccessReadDir           = 1 << 3
	AccessRemoveDir         = 1 << 4
	AccessRemoveFile        = 1 << 5
	AccessMakeChar          = 1 << 6
	AccessMakeDir           = 1 << 7
	AccessMakeReg           = 1 << 8
	AccessMakeSock          = 1 << 9
	AccessMakeFifo          = 1 << 10
	AccessMakeBlock         = 1 << 11
	AccessMakeSym           = 1 << 12
	RuleTypePathBeneath     = 1
	LandlockRulesetAttrSize = unsafe.Sizeof(landlockRulesetAttr{})
)

const landlockABI = 0 // Version 1 ABI

// ---------- Structs ----------

type landlockPathBeneathAttr struct {
	AllowedAccess uint64
	ParentFd      uint32
	_             [4]byte // required padding
}

// ---------- Syscall wrappers ----------

// Create a new ruleset and return its FD
func CreateRuleset(accessMask uint64) (int, error) {
	attr := landlockRulesetAttr{
		HandledAccessFs: accessMask,
	}

	fd, _, errno := unix.Syscall6(
		unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)),
		uintptr(LandlockRulesetAttrSize),
		uintptr(landlockABI),
		0, 0, 0,
	)
	if errno != 0 {
		return -1, fmt.Errorf("CreateRuleset syscall failed: %w", errno)
	}
	return int(fd), nil
}

// Add a rule allowing access to a specific path
func AddPathRule(rulesetFd int, path string, accessMask uint64) error {
	pathFd, err := unix.Open(path, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("failed to open path %s: %w", path, err)
	}
	defer unix.Close(pathFd)

	attr := landlockPathBeneathAttr{
		AllowedAccess: accessMask,
		ParentFd:      uint32(pathFd),
	}

	_, _, errno := unix.Syscall6(
		unix.SYS_LANDLOCK_ADD_RULE,
		uintptr(rulesetFd),
		uintptr(RuleTypePathBeneath),
		uintptr(unsafe.Pointer(&attr)),
		0, 0, 0,
	)
	if errno != 0 {
		return fmt.Errorf("AddPathRule syscall failed: %w", errno)
	}
	return nil
}

// Enforce the ruleset on current thread and its future children
func RestrictSelf(rulesetFd int) error {
	// Need to call prctl first to set NO_NEW_PRIVS
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("prctl PR_SET_NO_NEW_PRIVS failed: %w", err)
	}

	_, _, errno := unix.Syscall(
		unix.SYS_LANDLOCK_RESTRICT_SELF,
		uintptr(rulesetFd),
		0,
		0,
	)
	if errno != 0 {
		return fmt.Errorf("RestrictSelf syscall failed: %w", errno)
	}

	return nil
}

func CloseFd(fd int) {
	unix.Close(fd)
}
