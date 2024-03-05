package cmd

import (
	"log"

	"github.com/bwiggs/spacetraders-go/client"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates local data for systems, markets and shipyards.",
	Run: func(cmd *cobra.Command, args []string) {
		err := updateSystemData("X1-HK42")
		if err != nil {
			log.Fatal(err)
		}
	},
}

func updateSystemData(system string) error {
	client, err := client.Client()
	if err != nil {
		return err
	}

	r, err := repo.GetRepo()
	if err != nil {
		return err
	}

	err = tasks.ScanSystem(client, r, system)
	if err != nil {
		return err
	}

	return nil
}
