//go:build !windows

package etchosts

import (
	"errors"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

func applyPlatformOwnership(tmpPath string) {
	if err := os.Chown(tmpPath, os.Getuid(), os.Getgid()); err != nil {
		var pe *fs.PathError
		if errors.As(err, &pe) && errors.Is(pe.Err, unix.EPERM) {
			// Expected when unprivileged or running under sudo
		}
	}
}
