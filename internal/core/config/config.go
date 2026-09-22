package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/orekasep/go-hosts-cli/internal/domain"
	"gopkg.in/yaml.v3"
)

// DefaultConfigDirName is the directory inside the user's home folder.
const DefaultConfigDirName = ".hostcli"

// DefaultConfigFileName is the file name for storing the YAML configuration.
const DefaultConfigFileName = "hosts.yaml"

// GetUserConfigDir returns the directory path ~/.hostcli/ (or %USERPROFILE%\.hostcli\ on Windows).
func GetUserConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	return filepath.Join(home, DefaultConfigDirName), nil
}

// GetUserConfigPath returns the full path to ~/.hostcli/hosts.yaml.
func GetUserConfigPath() (string, error) {
	dir, err := GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DefaultConfigFileName), nil
}

// GetSystemHostsPath returns the standard system hosts path depending on OS.
func GetSystemHostsPath() string {
	if runtime.GOOS == "windows" {
		sysRoot := os.Getenv("SystemRoot")
		if sysRoot == "" {
			sysRoot = `C:\Windows`
		}
		return filepath.Join(sysRoot, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// LoadConfig reads and parses the YAML hosts file from path.
// If the file does not exist, an empty initialized HostsFile is returned.
func LoadConfig(path string) (*domain.HostsFile, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &domain.HostsFile{
			Version: 1,
			Groups:  []domain.Group{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var hf domain.HostsFile
	if err := yaml.Unmarshal(data, &hf); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	if hf.Version == 0 {
		hf.Version = 1
	}

	return &hf, nil
}

// SaveConfig serializes the HostsFile to YAML and writes it atomically to path.
func SaveConfig(path string, hf *domain.HostsFile) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(hf)
	if err != nil {
		return fmt.Errorf("failed to serialize hosts config to YAML: %w", err)
	}

	// Write to temporary file first then rename for atomic write in user home
	tmpFile, err := os.CreateTemp(dir, ".hosts-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to replace config file %s: %w", path, err)
	}

	return nil
}
