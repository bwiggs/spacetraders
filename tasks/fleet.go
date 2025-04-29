package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
)

func UpdateFleet(client api.Invoker, repo *repo.Repo) error {

	page := 1
	limit := 20
	for {
		slog.Info(fmt.Sprintf("updating ships: page %d", page), "page", page)
		res, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{Page: api.NewOptInt(page), Limit: api.NewOptInt(limit)})
		if err != nil {
			return err
		}

		if len(res.Data) == 0 {
			break
		}

		err = repo.UpsertFleet(res.Data)
		if err != nil {
			return err
		}

		if len(res.Data) < limit {
			break
		}

		time.Sleep(500 * time.Millisecond)

		page++
	}

	return nil
}
