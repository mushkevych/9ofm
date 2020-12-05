package utils

import (
	"os"
	"os/exec"
	"strings"
)

// CleanArgs trims the whitespace from the given set of strings.
func CleanArgs(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, strings.Trim(str, " "))
		}
	}
	return r
}

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
