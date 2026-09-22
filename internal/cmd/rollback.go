package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/orekasep/go-hosts-cli/internal/apply"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/spf13/cobra"
)

var (
	rollbackOriginal bool
	rollbackDryRun   bool
)

func newRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Revert system hosts file changes",
		Long: `Removes the hostcli managed block from the system hosts file,
or restores the pristine pre-migration backup using --original.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sysHostsPath := config.GetSystemHostsPath()
			priv := apply.CheckPrivileges(sysHostsPath)

			if !rollbackDryRun && !priv.HasDirectWrite && !priv.IsElevated {
				if priv.CanSudo {
					fmt.Println("🔒 Permission required to write system hosts. Requesting sudo...")
					exe, err := os.Executable()
					if err != nil {
						exe = "hostcli"
					}
					cliArgs := []string{"rollback"}
					if rollbackOriginal {
						cliArgs = append(cliArgs, "--original")
					}
					c := exec.Command("sudo", append([]string{exe}, cliArgs...)...)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					return c.Run()
				}
				return fmt.Errorf("permission denied writing to %s (and sudo is unavailable)", sysHostsPath)
			}

			res, err := apply.Rollback(rollbackOriginal, rollbackDryRun)
			if err != nil {
				return err
			}

			fmt.Printf("✓ %s\n", res.Message)
			return nil
		},
	}

	cmd.Flags().BoolVar(&rollbackOriginal, "original", false, "restore from original pre-migration backup (hosts.original.bak)")
	cmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "preview rollback without modifying files")
	return cmd
}
