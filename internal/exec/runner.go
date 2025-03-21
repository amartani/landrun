package exec

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/zouuup/landrun/internal/log"
)

func Run(args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	log.Info("Executing: %v", args)

	// Replace current process image with the target command
	return syscall.Exec(binary, args, os.Environ())
}
