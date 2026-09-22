package apply

import (
	"fmt"
	"os/exec"
	"runtime"
)

// FlushDNSCache flushes the operating system DNS resolver cache.
func FlushDNSCache() error {
	switch runtime.GOOS {
	case "darwin":
		// macOS
		_ = exec.Command("dscacheutil", "-flushcache").Run()
		return exec.Command("killall", "-HUP", "mDNSResponder").Run()

	case "linux":
		// Try resolvectl first (modern systemd)
		if err := exec.Command("resolvectl", "flush-caches").Run(); err == nil {
			return nil
		}
		// Fallback to nscd if present
		return exec.Command("systemctl", "restart", "nscd").Run()

	case "windows":
		// Windows
		return exec.Command("ipconfig", "/flushdns").Run()

	default:
		return fmt.Errorf("unsupported platform for automatic DNS flush: %s", runtime.GOOS)
	}
}
