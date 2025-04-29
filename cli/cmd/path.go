package cmd

import (
	algos "github.com/bwiggs/spacetraders-go/algos/routing"
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(pathCmd)
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "dump best path for ship from origin to destination",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := repo.GetRepo()
		if err != nil {
			panic(err)
		}

		waypoints, _ := r.GetWaypoints(viper.GetString("system"))
		ship := &api.Ship{
			Symbol: "BWIGGS-1",
			Nav: api.ShipNav{FlightMode: api.ShipNavFlightModeCRUISE,
				WaypointSymbol: api.WaypointSymbol(args[0]),
			},
			Fuel: api.ShipFuel{
				Current:  400,
				Capacity: 400,
			},
		}
		cost, path := algos.FindPath(ship, args[1], waypoints)
		spew.Dump(cost, path)
	},
}
