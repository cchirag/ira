//go:build windows
// +build windows

package spawn

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func binaryName() string {
	return "irad.exe"
}

func spawn(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
		HideWindow:    true,
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}
