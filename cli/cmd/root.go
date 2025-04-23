package cmd

import (
	"fmt"
	"os"

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
	// kernel, err := kernel.New()
	// if err != nil {
	// 	slog.Error("failed to create kernel", "err", err)
	// 	return
	// }

	// kernel.Run()

	// select {}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
