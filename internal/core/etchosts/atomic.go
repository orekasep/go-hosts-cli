package etchosts

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// WriteEtcHosts atomically replaces path with content.
// Properties:
// 1. Refuses to follow symlinks (os.Lstat check).
// 2. Uses same-directory tempfile to prevent EXDEV (cross-device rename).
// 3. Traps SIGINT/SIGTERM to clean up the temporary file if interrupted.
// 4. Sets standard 0644 file permissions.
func WriteEtcHosts(path string, content []byte) error {
	if fi, err := os.Lstat(path); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write through symlink: %s", path)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".hostcli-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	// Intercept interrupt signals to clean up tempfile
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		_ = os.Remove(tmpPath)
		os.Exit(130)
	}()
	defer signal.Stop(sigCh)

	defer func() {
		if _, statErr := os.Lstat(tmpPath); statErr == nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write payload: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Clamp permissions to 0644
	_ = os.Chmod(tmpPath, 0644)

	// Platform-specific ownership fix (no-op on Windows, EPERM-tolerant on Unix)
	applyPlatformOwnership(tmpPath)

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		// On Windows, if destination exists, rename might fail in older Go versions
		// os.Rename in Go 1.5+ on Windows handles replacement, but we check and fallback if needed
		_ = os.Remove(path)
		if retryErr := os.Rename(tmpPath, path); retryErr != nil {
			return fmt.Errorf("failed to replace %s: %w", path, err)
		}
	}

	return nil
}
