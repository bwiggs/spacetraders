package bot

import (
	"context"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/actors"
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
)

func Start(client *api.Client, r *repo.Repo) {
	ships, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{})
	if err != nil {
		slog.Error("bot failed to load ships")
	}

	fleet := make(map[string]*api.Ship)
	harvestors := []*api.Ship{}
	for _, s := range ships.Data {
		fleet[s.Symbol] = &s
		if s.Registration.Role == api.ShipRoleEXCAVATOR || s.Registration.Role == api.ShipRoleCOMMAND {
			harvestors = append(harvestors, &s)
		}
	}

	// s := fleet["BWIGGS-1"]
	// harvester := actors.NewShip(s, client)
	// harvester.SetMission(actors.NewTradeMission("EQUIPMENT", "X1-HK42-K80", "X1-HK42-A1"))

	// excavator := actors.NewShip(fleet["BWIGGS-5"], client)
	// excavator.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))

	for _, s := range harvestors {
		harvester := actors.NewShip(s, client)
		harvester.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))
	}
}
