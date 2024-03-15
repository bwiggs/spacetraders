package bot

import (
	"context"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/actors"
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/go-faster/errors"
)

func Start(client *api.Client, r *repo.Repo) {

	ships, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{Limit: api.NewOptInt(20)})
	if err != nil {
		slog.Error(errors.Wrap(err, "bot failed to load ships").Error())
		return
	}

	fleet := make(map[string]*api.Ship)
	fleetByType := make(map[string][]*api.Ship)
	for _, s := range ships.Data {
		fleet[s.Symbol] = &s
		role := string(s.Registration.Role)

		if _, found := fleetByType[role]; !found {
			fleetByType[role] = []*api.Ship{}
		}

		fleetByType[role] = append(fleetByType[role], &s)
	}

	// commandShip := actors.NewShip(fleet["BWIGGS-1"], client)
	// contracts, err := client.GetContracts(context.TODO(), api.GetContractsParams{})
	// if err != nil {
	// 	slog.Error("failed to load contracts")
	// }
	// for _, c := range contracts.Data {
	// 	if c.Accepted && !c.Fulfilled {
	// 		commandShip.SetMission(actors.NewContractMission(client, r, &c))
	// 		break
	// 	}
	// }

	extractMission := actors.NewExtractionMission(client, r, "X1-QY42-CZ5F")
	command := actors.NewShip(fleet["BWIGGS-1"], client)
	extractMission.AssignShip(actors.MissionShipRoleExcavatorTransporter, command)
	for _, s := range fleetByType["EXCAVATOR"] {
		ship := actors.NewShip(s, client)
		extractMission.AssignShip(actors.MissionShipRoleExcavator, ship)
	}
	for _, s := range fleetByType["TRANSPORT"] {
		ship := actors.NewShip(s, client)
		extractMission.AssignShip(actors.MissionShipRoleTransporter, ship)
	}

	// s := fleet["BWIGGS-1"]
	// harvester := actors.NewShip(s, client)
	// harvester.SetMission(actors.NewTradeMission("EQUIPMENT", "X1-HK42-K80", "X1-HK42-A1"))

	// excavator := actors.NewShip(fleet["BWIGGS-5"], client)
	// excavator.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))

	// for _, s := range harvestors {
	// 	harvester := actors.NewShip(s, client)
	// 	harvester.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))
	// }
}
