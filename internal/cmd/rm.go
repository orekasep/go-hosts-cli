package cmd

import (
	"fmt"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/domain"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <hostname>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a host entry from configuration",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hostname := args[0]

			hf, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}

			found := false
			for i := range hf.Groups {
				var filtered []domain.HostEntry
				for _, e := range hf.Groups[i].Entries {
					if e.Hostname == hostname {
						found = true
					} else {
						filtered = append(filtered, e)
					}
				}
				hf.Groups[i].Entries = filtered
			}

			if !found {
				return fmt.Errorf("host entry %q not found", hostname)
			}

			if err := config.SaveConfig(cfgPath, hf); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Removed %q from configuration.\n", hostname)
			fmt.Println("  Run 'hostcli apply' to sync changes with system hosts.")
			return nil
		},
	}
}
