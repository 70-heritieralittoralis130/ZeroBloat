//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// prepareCmd hides the console window on Windows
func prepareCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
