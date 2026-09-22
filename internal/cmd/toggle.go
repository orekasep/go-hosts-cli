package cmd

import (
	"fmt"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/spf13/cobra"
)

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <hostname>",
		Short: "Enable a host entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setEntryEnabled(args[0], true)
		},
	}
}

func newDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <hostname>",
		Short: "Disable a host entry without deleting it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setEntryEnabled(args[0], false)
		},
	}
}

func setEntryEnabled(hostname string, enabled bool) error {
	hf, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	entry, grp := hf.FindEntry(hostname)
	if entry == nil {
		return fmt.Errorf("host entry %q not found", hostname)
	}

	entry.Enabled = enabled

	if err := config.SaveConfig(cfgPath, hf); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	state := "enabled"
	if !enabled {
		state = "disabled"
	}

	fmt.Printf("✓ Host %q in group %q is now %s.\n", hostname, grp, state)
	fmt.Println("  Run 'hostcli apply' to update system hosts.")
	return nil
}
