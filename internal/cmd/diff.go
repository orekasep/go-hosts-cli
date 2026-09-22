package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/core/etchosts"
	"github.com/orekasep/go-hosts-cli/internal/core/render"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show differences between configuration and system hosts file",
		RunE: func(cmd *cobra.Command, args []string) error {
			hf, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}

			sysHostsPath := config.GetSystemHostsPath()
			currentContent, err := os.ReadFile(sysHostsPath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to read system hosts: %w", err)
			}

			renderedBlock := render.RenderManagedBlock(hf)
			newContent, err := etchosts.ReplaceManagedBlock(currentContent, []byte(renderedBlock))
			if err != nil {
				return fmt.Errorf("failed to merge block: %w", err)
			}

			if bytes.Equal(currentContent, newContent) {
				fmt.Println("✓ System hosts file is already up to date with configuration. No differences.")
				return nil
			}

			fmt.Println("=== Diff: System Hosts vs hostcli Config ===")
			_, managed, _, _ := etchosts.ExtractManagedBlock(currentContent)
			if managed == nil {
				fmt.Println("Status: No hostcli block currently in system hosts.")
				fmt.Printf("Will append managed block with %d enabled entries.\n\n", hf.EnabledEntries())
			} else {
				fmt.Println("--- Current Managed Block in System Hosts ---")
				fmt.Print(string(managed))
				fmt.Println("---------------------------------------------")
			}

			fmt.Println("\n+++ Proposed New Managed Block +++")
			fmt.Print(renderedBlock)
			fmt.Println("++++++++++++++++++++++++++++++++++")
			return nil
		},
	}
}
