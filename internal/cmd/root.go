package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	Version = "1.0.0"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "hostcli",
	Short: "Modern cross-platform /etc/hosts manager with TUI and CLI",
	Long: `hostcli is a modern /etc/hosts manager with a keyboard-driven Bubble Tea TUI
and a scriptable CLI, backed by a clean YAML source of truth at ~/.hostcli/hosts.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Launch TUI when invoked without subcommands
		p := tea.NewProgram(tui.InitialModel(cfgPath), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultCfg, _ := config.GetUserConfigPath()
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", defaultCfg, "path to YAML configuration file")

	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newRollbackCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newRmCmd())
	rootCmd.AddCommand(newEnableCmd())
	rootCmd.AddCommand(newDisableCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newVersionCmd())
}
