package cmd

import (
	"fmt"
	"os"

	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/core/migrate"
	"github.com/spf13/cobra"
)

var importFromPath string

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import existing system hosts into hostcli YAML config",
		Long:  "Parses the existing system hosts file (or a specified file) and generates ~/.hostcli/hosts.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			sourcePath := importFromPath
			if sourcePath == "" {
				sourcePath = config.GetSystemHostsPath()
			}

			fmt.Printf("Reading hosts file from: %s\n", sourcePath)
			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to read hosts file: %w", err)
			}

			hf, err := migrate.ParseHostsFile(content)
			if err != nil {
				return fmt.Errorf("failed to parse hosts file: %w", err)
			}

			if err := config.SaveConfig(cfgPath, hf); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Successfully imported %d entries across %d groups into %s\n",
				hf.TotalEntries(), len(hf.Groups), cfgPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&importFromPath, "from", "", "path to hosts file to import (defaults to system hosts)")
	return cmd
}
