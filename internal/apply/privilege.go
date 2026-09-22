package apply

import (
	"os"
	"os/exec"
	"runtime"
)

// PrivilegeStatus captures the current execution permission environment.
type PrivilegeStatus struct {
	HasDirectWrite bool
	CanSudo        bool
	IsElevated     bool
}

// CheckPrivileges checks whether the process can directly write to targetPath
// or whether elevation (sudo or Windows admin) is required.
func CheckPrivileges(targetPath string) PrivilegeStatus {
	// Test direct write permission
	canWrite := CanWriteTarget(targetPath)
	if canWrite {
		return PrivilegeStatus{
			HasDirectWrite: true,
			CanSudo:        true,
			IsElevated:     true,
		}
	}

	if runtime.GOOS == "windows" {
		elevated := isWindowsAdmin()
		return PrivilegeStatus{
			HasDirectWrite: elevated,
			CanSudo:        false,
			IsElevated:     elevated,
		}
	}

	// macOS / Linux: check if sudo binary is present
	_, sudoErr := exec.LookPath("sudo")
	return PrivilegeStatus{
		HasDirectWrite: false,
		CanSudo:        sudoErr == nil,
		IsElevated:     os.Geteuid() == 0,
	}
}

// CanWriteTarget tests whether targetPath is writable by the current process.
func CanWriteTarget(targetPath string) bool {
	f, err := os.OpenFile(targetPath, os.O_WRONLY, 0644)
	if err == nil {
		_ = f.Close()
		return true
	}
	return false
}

func isWindowsAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// On Windows, running "net session" or checking access to System32 tests elevation
	cmd := exec.Command("net", "session")
	return cmd.Run() == nil
}
