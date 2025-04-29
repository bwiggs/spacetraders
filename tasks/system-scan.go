package tasks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/go-faster/errors"
)

func ScanSystem(client api.Invoker, repo *repo.Repo, system string) error {
	log := slog.With("job", "ScanSystem", "system", system)

	var err error

	log.Info("ScanSystem: " + system)
	log.Info("ScanSystem: waypoints")
	err = ScanWaypoints(client, repo, system)
	if err != nil {
		return errors.Wrap(err, "ScanSystem: failed to scan waypoints")
	}

	log.Info("ScanSystem: markets")
	err = ScanMarkets(client, repo, system)
	if err != nil {
		return errors.Wrap(err, "ScanSystem: failed to scan markets")
	}

	log.Info("ScanSystem: shipyards")
	err = ScanShipyards(client, repo, system)
	if err != nil {
		return errors.Wrap(err, "ScanSystem: failed to scan shipyards")
	}

	log.Info("ScanSystem: done")

	return nil
}

func ScanWaypoints(client api.Invoker, repo *repo.Repo, system string) error {
	baselog := slog.With("job", "ScanWaypoints", "system", system)
	ctx := context.TODO()
	page := 1
	limit := 20
	for {
		log := baselog.With("page", page, "limit", limit)
		log.Info("fetching waypoints")
		params := api.GetSystemWaypointsParams{SystemSymbol: system, Limit: api.NewOptInt(limit), Page: api.NewOptInt(page)}
		res, err := client.GetSystemWaypoints(ctx, params)
		if err != nil {
			return errors.Wrap(err, "ScanWaypoints: failed get system waypoints")
		}

		if len(res.Data) == 0 {
			log.Info("done")
			break
		}

		log.Info(fmt.Sprintf("saving %d waypoints", len(res.Data)))
		err = repo.UpsertWaypoints(res.Data)
		if err != nil {
			return errors.Wrap(err, "ScanWaypoints: failed get upsert waypoints")
		}
		log.Info("page complete")
		page++
	}
	return nil
}

func ScanMarkets(client api.Invoker, repo *repo.Repo, system string) error {
	waypoints, err := repo.GetSystemWaypointsByTrait(system, "MARKETPLACE")
	if err != nil {
		return err
	}

	for _, wp := range waypoints {
		err := ScanMarket(client, repo, wp.Symbol)
		if err != nil {
			slog.Error(errors.Wrap(err, "failed to scan market").Error())
		}
	}

	return nil
}

func ScanMarket(client api.Invoker, repo *repo.Repo, wp string) error {
	slog.Debug("scanning market: " + wp)

	dat, err := client.GetMarket(context.TODO(), api.GetMarketParams{SystemSymbol: wp[:7], WaypointSymbol: wp})
	if err != nil {
		return errors.Wrap(err, "ScanMarket: failed get market")
	}

	err = repo.UpsertMarket(dat.Data)
	if err != nil {
		return errors.Wrap(err, "ScanMarket: failed upsert market")
	}
	return nil
}

func ScanShipyards(client api.Invoker, repo *repo.Repo, system string) error {
	waypoints, err := repo.GetSystemWaypointsByTrait(system, "SHIPYARD")
	if err != nil {
		return errors.Wrap(err, "ScanShipyards: failed to get waypoints")
	}

	if len(waypoints) == 0 {
		slog.Info("no shipyards found")
		return nil
	}

	for _, wp := range waypoints {
		ScanShipyard(client, repo, wp.Symbol)
	}

	return nil
}

func ScanShipyard(client api.Invoker, repo *repo.Repo, wp string) error {
	slog.Info("scanning shipyard: " + wp)

	dat, err := client.GetShipyard(context.TODO(), api.GetShipyardParams{SystemSymbol: wp[:7], WaypointSymbol: wp})
	if err != nil {
		return errors.Wrap(err, "ScanShipyard: failed to get shipyards")
	}

	err = repo.UpsertShipyard(dat.Data)
	if err != nil {
		return errors.Wrap(err, "ScanShipyard: failed to upsert shipyards")
	}
	return nil
}

func UpdateSystems(client api.Invoker, repo *repo.Repo) error {

	page := 1
	for {
		slog.Info(fmt.Sprintf("updating systems: page %d", page), "page", page)
		res, err := client.GetSystems(context.TODO(), api.GetSystemsParams{Limit: api.NewOptInt(20), Page: api.NewOptInt(page)})
		if err != nil {
			return errors.Wrap(err, "UpdateSystems: failed get systems")
		}

		if len(res.Data) == 0 {
			break
		}

		err = repo.UpsertSystems(res.Data)
		if err != nil {
			return errors.Wrap(err, "UpdateSystems: failed upsert systems")
		}

		page++
	}

	return nil
}
