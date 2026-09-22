package apply

import (
	"bytes"
	"fmt"
	"os"

	"github.com/orekasep/go-hosts-cli/internal/core/backup"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/core/etchosts"
	"github.com/orekasep/go-hosts-cli/internal/core/render"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

type ApplyResult struct {
	Changed    bool
	Message    string
	BackupPath string
	Preview    string
}

type RollbackResult struct {
	Changed          bool
	Message          string
	RestoredOriginal bool
}

// Apply updates the system hosts file with the enabled entries from hf.
func Apply(hf *domain.HostsFile, dryRun bool) (*ApplyResult, error) {
	sysHostsPath := config.GetSystemHostsPath()

	// 1. Ensure a pristine baseline backup exists
	if !dryRun {
		_ = backup.EnsureOriginalBackup(sysHostsPath)
	}

	// 2. Render the managed block
	renderedBlock := render.RenderManagedBlock(hf)

	// 3. Read current system hosts content
	currentContent, err := os.ReadFile(sysHostsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read system hosts (%s): %w", sysHostsPath, err)
	}

	// 4. Merge rendered block with current system hosts
	newContent, err := etchosts.ReplaceManagedBlock(currentContent, []byte(renderedBlock))
	if err != nil {
		return nil, fmt.Errorf("failed to merge managed block: %w", err)
	}

	// Check if already up-to-date
	if bytes.Equal(currentContent, newContent) {
		return &ApplyResult{
			Changed: false,
			Message: "System hosts is already up to date. No changes needed.",
		}, nil
	}

	if dryRun {
		return &ApplyResult{
			Changed: true,
			Message: "Dry-run: changes preview generated (nothing written).",
			Preview: string(newContent),
		}, nil
	}

	// 5. Create timestamped snapshot
	backupPath, err := backup.CreateTimestampBackup(sysHostsPath)
	if err != nil {
		return nil, fmt.Errorf("pre-apply backup failed: %w", err)
	}

	// 6. Write atomically to system hosts
	if err := etchosts.WriteEtcHosts(sysHostsPath, newContent); err != nil {
		return nil, fmt.Errorf("failed to write system hosts: %w", err)
	}

	// 7. Flush DNS cache
	_ = FlushDNSCache()

	return &ApplyResult{
		Changed:    true,
		Message:    fmt.Sprintf("Successfully applied %d enabled entries to %s", hf.EnabledEntries(), sysHostsPath),
		BackupPath: backupPath,
	}, nil
}

// Rollback reverts changes made to the system hosts file.
func Rollback(restoreOriginal bool, dryRun bool) (*RollbackResult, error) {
	sysHostsPath := config.GetSystemHostsPath()

	if restoreOriginal {
		if dryRun {
			return &RollbackResult{
				Changed:          true,
				Message:          "Dry-run: would restore system hosts from original backup.",
				RestoredOriginal: true,
			}, nil
		}

		// Snapshot before restoring
		_, _ = backup.CreateTimestampBackup(sysHostsPath)

		if err := backup.RestoreOriginal(sysHostsPath); err != nil {
			return nil, fmt.Errorf("failed to restore original backup: %w", err)
		}

		_ = FlushDNSCache()
		return &RollbackResult{
			Changed:          true,
			Message:          "Successfully restored system hosts to original pre-migration state.",
			RestoredOriginal: true,
		}, nil
	}

	// Default rollback: strip hostcli block cleanly
	currentContent, err := os.ReadFile(sysHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", sysHostsPath, err)
	}

	stripped, err := etchosts.StripManagedBlock(currentContent)
	if err != nil {
		return nil, fmt.Errorf("failed to strip managed block: %w", err)
	}

	if bytes.Equal(currentContent, stripped) {
		return &RollbackResult{
			Changed:          false,
			Message:          "No hostcli managed block found in system hosts. Nothing to rollback.",
			RestoredOriginal: false,
		}, nil
	}

	if dryRun {
		return &RollbackResult{
			Changed:          true,
			Message:          "Dry-run: would strip managed block from system hosts.",
			RestoredOriginal: false,
		}, nil
	}

	// Snapshot before stripping
	_, _ = backup.CreateTimestampBackup(sysHostsPath)

	if err := etchosts.WriteEtcHosts(sysHostsPath, stripped); err != nil {
		return nil, fmt.Errorf("failed to write system hosts: %w", err)
	}

	_ = FlushDNSCache()

	return &RollbackResult{
		Changed:          true,
		Message:          "Successfully removed hostcli block from system hosts.",
		RestoredOriginal: false,
	}, nil
}
