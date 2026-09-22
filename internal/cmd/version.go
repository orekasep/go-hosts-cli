package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hostcli version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hostcli version %s (commit: %s, date: %s)\n", Version, Commit, Date)
		},
	}
}
