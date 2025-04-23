package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/actors"
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/go-faster/errors"
)

func Start(client *api.Client, r *repo.Repo) {
	ships := []api.Ship{}
	page := 1
	processed := 0
	for {
		slog.Info(fmt.Sprintf("loading page %d", page))
		res, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{Limit: api.NewOptInt(20), Page: api.NewOptInt(page)})
		if err != nil {
			slog.Error(errors.Wrap(err, "bot failed to load ships").Error())
			return
		}
		ships = append(ships, res.Data...)
		processed += len(res.Data)
		if len(res.Data) == 0 {
			break
		}

		if processed == res.Meta.Total {
			break
		}
		page++
		time.Sleep(1 * time.Second)
	}

	fleet := make(map[string]*api.Ship)
	fleetByType := make(map[string][]*api.Ship)
	for _, s := range ships {
		fleet[s.Symbol] = &s
		role := string(s.Registration.Role)

		if _, found := fleetByType[role]; !found {
			fleetByType[role] = []*api.Ship{}
		}

		fleetByType[role] = append(fleetByType[role], &s)
	}

	contractMission(client, r, fleet)
	// tradeMission(client, r, fleet)
	// miningMission(client, r, fleet)
	// extractionMission(client, r, fleet)
}

func tradeMission(client *api.Client, r *repo.Repo, fleet map[string]*api.Ship) {
	commandShip := actors.NewShip(fleet["BWIGGS-1"], client)

	tradeMission := actors.NewTradeMission(client, r)
	tradeMission.AssignShip(actors.MissionShipRoleTrader, commandShip)

	// for _, s := range fleetByType[string(api.ShipRoleTRANSPORT)] {
	// 	ship := actors.NewShip(s, client)
	// 	tradeMission.AssignShip(actors.MissionShipRoleTrader, ship)
	// }
}
func miningMission(client *api.Client, r *repo.Repo, fleet map[string]*api.Ship) {
	// excavator := actors.NewShip(fleet["BWIGGS-5"], client)
	// excavator.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))

	// for _, s := range harvestors {
	// 	harvester := actors.NewShip(s, client)
	// 	harvester.SetMission(actors.NewMiningMission(r, "X1-HK42-AC5C"))
	// }
}

func extractionMission(client *api.Client, r *repo.Repo, fleet map[string]*api.Ship) {
	// // extract mission
	// {
	// 	extractMission := actors.NewExtractionMission(client, r, "X1-QY42-CZ5F")
	// 	// extractMission.AssignShip(actors.MissionShipRoleTransporter, command)
	// 	for _, s := range fleetByType[string(api.ShipRoleEXCAVATOR)] {
	// 		ship := actors.NewShip(s, client)
	// 		extractMission.AssignShip(actors.MissionShipRoleExcavator, ship)
	// 	}
	// 	for _, s := range fleetByType[string(api.ShipRoleHAULER)] {
	// 		ship := actors.NewShip(s, client)
	// 		extractMission.AssignShip(actors.MissionShipRoleHauler, ship)
	// 	}
	// }

	// // for _, s := range fleetByType[string(api.ShipRoleSURVEYOR)] {
	// // 	ship := actors.NewShip(s, client)
	// // 	extractMission.AssignShip(actors.MissionShipRoleSurveyor, ship)
	// // }
}

func contractMission(client *api.Client, r *repo.Repo, fleet map[string]*api.Ship) {
	commandShip := actors.NewShip(fleet["BWIGGS-1"], client)
	commandShip.SetMission(actors.NewContractMission(client, r))
}
