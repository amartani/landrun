package landlock

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// This is a stub. Replace with real syscall constants and logic later.
func IsSupported() bool {
	// TODO: actually detect kernel + CONFIG_LANDLOCK, etc.
	var uname unix.Utsname
	err := unix.Uname(&uname)
	return err == nil
}

// TODO: implement CreateRuleset, AddRule, RestrictSelf using raw syscalls

func CreateRuleset() error {
	// Stub
	return fmt.Errorf("not implemented")
}
