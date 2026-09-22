package cmd

import (
	"fmt"
	"strings"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/domain"
	"github.com/spf13/cobra"
)

var (
	addGroup    string
	addAliases  string
	addComment  string
	addDisabled bool
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <ip> <hostname>",
		Short: "Add a new host entry to the configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ip := args[0]
			hostname := args[1]

			if err := domain.ValidateIP(ip); err != nil {
				return err
			}
			if err := domain.ValidateHostname(hostname); err != nil {
				return err
			}

			hf, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}

			// Parse aliases
			var aliases []string
			if addAliases != "" {
				for _, a := range strings.Split(addAliases, ",") {
					trimmed := strings.TrimSpace(a)
					if trimmed != "" {
						if err := domain.ValidateHostname(trimmed); err != nil {
							return fmt.Errorf("invalid alias %q: %w", trimmed, err)
						}
						aliases = append(aliases, trimmed)
					}
				}
			}

			groupName := addGroup
			if groupName == "" {
				groupName = "default"
			}

			entry := domain.HostEntry{
				ID:       domain.NewULID(),
				IP:       ip,
				Hostname: hostname,
				Aliases:  aliases,
				Enabled:  !addDisabled,
				Comment:  addComment,
			}

			grp := hf.GetOrCreateGroup(groupName)
			grp.Entries = append(grp.Entries, entry)

			if err := config.SaveConfig(cfgPath, hf); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Added '%s' -> %s (Group: %s) to %s\n", hostname, ip, groupName, cfgPath)
			fmt.Println("  Run 'hostcli apply' to write changes to system hosts.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&addGroup, "group", "g", "default", "group to place the entry in")
	cmd.Flags().StringVarP(&addAliases, "alias", "a", "", "comma-separated aliases")
	cmd.Flags().StringVarP(&addComment, "comment", "m", "", "optional comment description")
	cmd.Flags().BoolVar(&addDisabled, "disabled", false, "add entry in disabled state")

	return cmd
}
