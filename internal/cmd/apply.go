package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/orekasep/go-hosts-cli/internal/apply"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/spf13/cobra"
)

var applyDryRun bool

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply enabled hosts entries to the system hosts file",
		Long:  "Renders all enabled entries from ~/.hostcli/hosts.yaml into the system hosts file between boundary markers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			hf, err := config.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			sysHostsPath := config.GetSystemHostsPath()
			priv := apply.CheckPrivileges(sysHostsPath)

			// If not dry-run and not writable, check sudo
			if !applyDryRun && !priv.HasDirectWrite && !priv.IsElevated {
				if priv.CanSudo {
					fmt.Println("🔒 Permission required to write system hosts. Requesting sudo...")
					exe, err := os.Executable()
					if err != nil {
						exe = "hostcli"
					}
					c := exec.Command("sudo", exe, "apply", "--config", cfgPath)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					return c.Run()
				}
				return fmt.Errorf("permission denied writing to %s (and sudo is unavailable)", sysHostsPath)
			}

			result, err := apply.Apply(hf, applyDryRun)
			if err != nil {
				return err
			}

			if applyDryRun {
				fmt.Println("--- DRY RUN PREVIEW (would be written to system hosts) ---")
				fmt.Println(result.Preview)
				fmt.Println("----------------------------------------------------------")
				fmt.Println("✓ Dry-run completed. No files were modified.")
				return nil
			}

			fmt.Printf("✓ %s\n", result.Message)
			if result.BackupPath != "" {
				fmt.Printf("  Snapshot saved to: %s\n", result.BackupPath)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "preview changes without writing to system hosts")
	return cmd
}
