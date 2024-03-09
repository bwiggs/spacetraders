package bot

import (
	"context"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/actors"
	"github.com/bwiggs/spacetraders-go/api"
)

func Start(client *api.Client) {
	ships, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{})
	if err != nil {
		slog.Error("bot failed to load ships")
	}

	fleet := make(map[string]*api.Ship)
	for _, s := range ships.Data {
		fleet[s.Symbol] = &s
	}

	s := fleet["BWIGGS-1"]
	harvester := actors.NewShip(s, client)
	harvester.SetMission(actors.NewTradeMission("CLOTHING", "X1-HK42-K80", "X1-HK42-A1"))
}
