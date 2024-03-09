package cmd

import (
	"fmt"
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
		client, err := client.Client()
		if err != nil {
			log.Fatal(err)
		}

		r, err := repo.GetRepo()
		if err != nil {
			log.Fatal(err)
		}

		nargs := len(args)
		if nargs < 2 {
			fmt.Println("arg expected: all, markets, shipyards, waypoints")
			return
		}

		target := args[1]

		switch args[0] {
		case "all":
			err = tasks.ScanSystem(client, r, target)
		case "markets":
			if len(target) == 7 {
				err = tasks.ScanMarkets(client, r, target)
			} else {
				err = tasks.ScanMarket(client, r, target)
			}
		case "shipyards":
			err = tasks.ScanShipyards(client, r, target)
		case "waypoints":
			err = tasks.ScanWaypoints(client, r, target)
		}

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
