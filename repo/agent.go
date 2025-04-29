package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repo) UpsertAgents(agents []api.Agent) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertAgents, err := tx.Prepare("INSERT OR REPLACE INTO agents (symbol, faction, headquarters, credits, shipCount, json) values (?, ?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer upsertAgents.Close()

	for _, a := range agents {

		buf, err := a.MarshalJSON()
		if err != nil {
			return err
		}

		_, err = upsertAgents.Exec(
			a.GetSymbol(),
			a.GetStartingFaction(),
			a.GetHeadquarters(),
			a.GetCredits(),
			a.GetShipCount(),
			buf,
		)
		if err != nil {
			panic(err)
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) GetAgents() ([]*api.Agent, error) {
	bufs := [][]byte{}
	if err := r.db.Select(&bufs, "SELECT json from agents"); err != nil {
		return nil, err
	}

	agents := make([]*api.Agent, len(bufs))
	for i := range bufs {
		a := &api.Agent{}
		if err := a.UnmarshalJSON(bufs[i]); err != nil {
			return nil, err
		}
		agents[i] = a
	}
	return agents, nil
}
