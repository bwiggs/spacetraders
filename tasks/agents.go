package tasks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/go-faster/errors"
)

func UpdateAgents(client api.Invoker, repo *repo.Repo) error {
	log := slog.With("job", "UpdateAgents")

	ctx := context.TODO()
	page := 1
	limit := 20
	for {
		log.Info(fmt.Sprintf("updating agents: page %d", page), "page", page)
		res, err := client.GetAgents(ctx, api.GetAgentsParams{Page: api.NewOptInt(page), Limit: api.NewOptInt(limit)})
		if err != nil {
			return errors.Wrap(err, "UpdateAgents: failed get agents")
		}
		if len(res.Data) == 0 || len(res.Data) < limit {
			log.Info("done")
			break
		}

		if err := repo.UpsertAgents(res.Data); err != nil {
			return errors.Wrap(err, "UpdateAgents: failed upsert agents")
		}

		page++

	}

	return nil
}
