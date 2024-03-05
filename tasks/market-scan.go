package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
)

func ScanSystem(client *api.Client, repo *repo.Repo, system string) error {
	log := slog.With("job", "ScanSystem", "system", system)
	err := ScanWaypoints(client, repo, system)
	if err != nil {
		return err
	}

	log.Info("scanning markets")
	err = ScanGoods(client, repo, system)
	if err != nil {
		return err
	}

	log.Info("done")

	return nil
}

func ScanWaypoints(client *api.Client, repo *repo.Repo, system string) error {
	log := slog.With("job", "ScanWaypoints", "system", system)
	ctx := context.TODO()
	page := 1
	limit := 20
	for {
		log = log.With("page", page, "limit", limit)
		log.Info("fetching systems")
		params := api.GetSystemWaypointsParams{SystemSymbol: system, Limit: api.NewOptInt(limit), Page: api.NewOptInt(page)}
		res, err := client.GetSystemWaypoints(ctx, params)
		if err != nil {
			return err
		}

		if len(res.Data) == 0 {
			log.Info("done")
			break
		}

		log.Info(fmt.Sprintf("saving %d waypoints", len(res.Data)))
		err = repo.UpsertWaypoints(res.Data)
		if err != nil {
			return err
		}
		log.Info("page complete")
		page++

	}
	return nil
}

func ScanGoods(client *api.Client, repo *repo.Repo, system string) error {
	waypoints, err := repo.GetSystemMarkets(system)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	for _, wp := range waypoints {
		slog.Info("scanning market: " + wp)

		dat, err := client.GetMarket(ctx, api.GetMarketParams{SystemSymbol: system, WaypointSymbol: wp})
		if err != nil {
			return err
		}

		err = repo.UpsertMarket(dat.Data)
		if err != nil {
			return err
		}

		time.Sleep(600 * time.Millisecond)
	}

	// waypoints, err = repo.GetSystemShipyards(system)
	// if err != nil {
	// 	return err
	// }

	// ctx = context.TODO()
	// for _, wp := range waypoints {
	// 	slog.Info("scanning shipyard: " + wp)

	// 	dat, err := client.GetShipyard(ctx, api.GetShipyardParams{SystemSymbol: system, WaypointSymbol: wp})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = repo.UpsertMarket(dat.Data)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	time.Sleep(600 * time.Millisecond)
	// }

	return nil
}
