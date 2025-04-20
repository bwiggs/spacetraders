package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
)

func UpdateFleet(client *api.Client, repo *repo.Repo) error {

	page := 1
	for {
		slog.Info(fmt.Sprintf("updating ships: page %d", page), "page", page)
		res, err := client.GetMyShips(context.TODO(), api.GetMyShipsParams{Page: api.NewOptInt(page), Limit: api.NewOptInt(20)})
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

		time.Sleep(500 * time.Millisecond)

		page++
	}

	return nil
}
