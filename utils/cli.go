package utils

import (
	"os"
	"os/exec"
)

// runCmd runs a given shell command in the current tty
func runCmd(cmdStr string, args ...string) error {
	allArgs := CleanArgs(args)

	cmd := exec.Command(cmdStr, allArgs...)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
