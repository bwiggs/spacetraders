package cmd

import (
	"log/slog"

	"github.com/bwiggs/spacetraders-go/kernel"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "st",
	Short: "spacetraders.io cli interface",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	kernel, err := kernel.New()
	if err != nil {
		slog.Error("failed to create kernel", "err", err)
		return
	}

	kernel.Run()

	select {}
}
