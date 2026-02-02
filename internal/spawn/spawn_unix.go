//go:build unix
// +build unix

package spawn

import (
	"os/exec"
	"syscall"
)

func binaryName() string {
	return "irad"
}

func spawn(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}
