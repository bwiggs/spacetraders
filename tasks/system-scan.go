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

	var err error

	log.Info("ScanSystem: waypoints")
	err = ScanWaypoints(client, repo, system)
	if err != nil {
		return err
	}

	log.Info("ScanSystem: markets")
	err = ScanMarkets(client, repo, system)
	if err != nil {
		return err
	}

	log.Info("ScanSystem: shipyards")
	err = ScanShipyards(client, repo, system)
	if err != nil {
		return err
	}

	log.Info("ScanSystem: done")

	return nil
}

func ScanWaypoints(client *api.Client, repo *repo.Repo, system string) error {
	baselog := slog.With("job", "ScanWaypoints", "system", system)
	ctx := context.TODO()
	page := 1
	limit := 20
	for {
		log := baselog.With("page", page, "limit", limit)
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

func ScanMarkets(client *api.Client, repo *repo.Repo, system string) error {
	waypoints, err := repo.GetSystemWaypointsByTrait(system, "MARKETPLACE")
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

	return nil
}

func ScanShipyards(client *api.Client, repo *repo.Repo, system string) error {
	waypoints, err := repo.GetSystemWaypointsByTrait(system, "SHIPYARD")
	if err != nil {
		return err
	}

	ctx := context.TODO()
	for _, wp := range waypoints {
		slog.Info("scanning shipyard: " + wp)

		dat, err := client.GetShipyard(ctx, api.GetShipyardParams{SystemSymbol: system, WaypointSymbol: wp})
		if err != nil {
			return err
		}

		err = repo.UpsertShipyard(dat.Data)
		if err != nil {
			return err
		}

		time.Sleep(600 * time.Millisecond)
	}

	return nil
}