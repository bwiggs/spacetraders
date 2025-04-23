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

		wps, _ := r.GetWaypoints(viper.GetString("system"))
		ship := &api.Ship{Symbol: "BWIGGS-1", Nav: api.ShipNav{WaypointSymbol: api.WaypointSymbol(args[0])}}
		path := algos.FindPath(ship, args[1], wps)
		spew.Dump(path)
	},
}
