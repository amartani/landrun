package landlock

import (
	"fmt"
	"unsafe"

	"github.com/zouuup/landrun/internal/log"
	"golang.org/x/sys/unix"
)

type landlockRulesetAttr struct {
	HandledAccessFs uint64
	_               [8]byte // padding: must match kernel struct size
}

func IsSupported() bool {
	attr := landlockRulesetAttr{
		HandledAccessFs: 0, // no rules, just probing
	}

	log.Debug("attempting Landlock syscall with attr: %+v", attr)

	_, _, errno := unix.Syscall6(
		unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
		0, 0, 0, 0,
	)

	log.Debug("Landlock syscall result: %s (errno: %d)", unix.ErrnoName(errno), errno)

	switch errno {
	case 0:
		log.Debug("Landlock is supported")
		return true
	case unix.ENOSYS, unix.EOPNOTSUPP:
		log.Error("Landlock not supported by kernel")
		return false
	case unix.EINVAL, unix.ENOMSG:
		log.Debug("Landlock syscall exists (probe successful)")
		return true
	default:
		errMsg := unix.ErrnoName(errno)
		if errMsg == "" {
			errMsg = fmt.Sprintf("unknown error %d", errno)
		}
		log.Error("unexpected Landlock error: %s (errno: %d)", errMsg, errno)
		return false
	}
}
