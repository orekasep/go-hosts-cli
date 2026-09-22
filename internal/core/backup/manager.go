package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/core/etchosts"
)

const (
	OriginalBackupFileName = "hosts.original.bak"
	BackupsSubDir          = "backups"
)

// GetBackupsDir returns the directory path for storing hosts backups (~/.hostcli/backups).
func GetBackupsDir() (string, error) {
	cfgDir, err := config.GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, BackupsSubDir), nil
}

// GetOriginalBackupPath returns the path to the original baseline hosts backup.
func GetOriginalBackupPath() (string, error) {
	bDir, err := GetBackupsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(bDir, OriginalBackupFileName), nil
}

// EnsureOriginalBackup copies the system hosts file to hosts.original.bak if not already present.
func EnsureOriginalBackup(systemHostsPath string) error {
	origPath, err := GetOriginalBackupPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(origPath); err == nil {
		return nil // Already backed up
	}

	data, err := os.ReadFile(systemHostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read system hosts for original backup: %w", err)
	}

	bDir := filepath.Dir(origPath)
	if err := os.MkdirAll(bDir, 0755); err != nil {
		return fmt.Errorf("failed to create backups directory: %w", err)
	}

	if err := os.WriteFile(origPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write original backup: %w", err)
	}

	return nil
}

// CreateTimestampBackup saves a timestamped snapshot before an apply or rollback.
func CreateTimestampBackup(systemHostsPath string) (string, error) {
	bDir, err := GetBackupsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(bDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backups directory: %w", err)
	}

	data, err := os.ReadFile(systemHostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read system hosts for snapshot: %w", err)
	}

	ts := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(bDir, fmt.Sprintf("hosts_%s.bak", ts))

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write timestamp backup: %w", err)
	}

	return backupPath, nil
}

// RestoreOriginal restores the system hosts file from the pristine original backup.
func RestoreOriginal(systemHostsPath string) error {
	origPath, err := GetOriginalBackupPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(origPath)
	if err != nil {
		return fmt.Errorf("original backup not found at %s: %w", origPath, err)
	}

	return etchosts.WriteEtcHosts(systemHostsPath, data)
}

// ListBackups returns all backup filenames sorted from newest to oldest.
func ListBackups() ([]string, error) {
	bDir, err := GetBackupsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(bDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".bak" {
			files = append(files, e.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, nil
}
