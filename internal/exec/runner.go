package exec

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func Run(args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	log.Printf("[landrun] Executing: %v", args)

	// Replace current process image with the target command
	return syscall.Exec(binary, args, os.Environ())
}
