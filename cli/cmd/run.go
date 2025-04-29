package cmd

import (
	"log/slog"

	"github.com/bwiggs/spacetraders-go/kernel"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "spacetraders.io cli interface",
	Run: func(cmd *cobra.Command, args []string) {
		kernel, err := kernel.New()
		if err != nil {
			slog.Error("failed to create kernel", "err", err)
			return
		}

		kernel.Run()

		select {}
	},
}
