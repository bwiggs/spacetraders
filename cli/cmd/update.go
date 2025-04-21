package cmd

import (
	"fmt"
	"log"

	"github.com/bwiggs/spacetraders-go/client"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		if nargs == 0 {
			fmt.Println("arg expected: all, market(s), shipyard(s), waypoint(s)")
			return
		}

		var target string
		if nargs == 2 {
			target = args[1]
		} else {
			target = viper.GetString("SYSTEM")
		}

		switch args[0] {
		case "all":
			fallthrough
		case "system":
			err = tasks.ScanSystem(client, r, target)
		case "markets":
			err = tasks.ScanMarkets(client, r, target)
		case "market":
			err = tasks.ScanMarket(client, r, target)
		case "shipyard":
			err = tasks.ScanShipyard(client, r, target)
		case "shipyards":
			err = tasks.ScanShipyards(client, r, target)
		case "systems":
			err = tasks.UpdateSystems(client, r)
		case "waypoints":
			err = tasks.ScanWaypoints(client, r, target)
		case "fleet":
			err = tasks.UpdateFleet(client, r)
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
