package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listAsJSON  bool
	listAsYAML  bool
	filterGroup string
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all configured host entries and groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			hf, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}

			if listAsJSON {
				data, err := json.MarshalIndent(hf, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if listAsYAML {
				data, err := yaml.Marshal(hf)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("Config File: %s\n", cfgPath)
			fmt.Printf("Total: %d entries (%d enabled)\n\n", hf.TotalEntries(), hf.EnabledEntries())

			for _, g := range hf.Groups {
				if filterGroup != "" && g.Name != filterGroup {
					continue
				}

				fmt.Printf("📁 Group: %s (%d entries)\n", g.Name, len(g.Entries))
				if len(g.Entries) == 0 {
					fmt.Println("   (No entries)")
					continue
				}

				fmt.Printf("   %-6s  %-16s  %-24s  %-15s  %s\n", "STATUS", "IP ADDRESS", "HOSTNAME", "ALIASES", "COMMENT")
				fmt.Println("   " + strings.Repeat("─", 80))

				for _, e := range g.Entries {
					status := "OFF"
					if e.Enabled {
						status = "ON "
					}
					aliasStr := strings.Join(e.Aliases, ",")
					if aliasStr == "" {
						aliasStr = "-"
					}
					fmt.Printf("   [%s]   %-16s  %-24s  %-15s  %s\n",
						status, e.IP, e.Hostname, aliasStr, e.Comment)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&listAsJSON, "json", false, "output as JSON format")
	cmd.Flags().BoolVar(&listAsYAML, "yaml", false, "output as raw YAML")
	cmd.Flags().StringVarP(&filterGroup, "group", "g", "", "filter entries by group name")

	return cmd
}
