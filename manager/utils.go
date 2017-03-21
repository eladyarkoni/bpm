package manager

import (
	"os"
	"syscall"
)

// IsProcessRunning checks if process is running
func IsProcessRunning(pid int) bool {
	osProc, osProcNotFound := os.FindProcess(pid)
	if osProcNotFound != nil {
		return false
	}
	if err := osProc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}
