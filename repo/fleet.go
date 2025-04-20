package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repo) UpsertFleet(ships []api.Ship) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertSystem, err := tx.Prepare("INSERT OR REPLACE INTO fleet (symbol, data) values (?, ?);")
	if err != nil {
		return err
	}
	defer upsertSystem.Close()

	for _, s := range ships {
		// upsert Market waypoints
		data, err := s.MarshalJSON()
		if err != nil {
			panic(err)
		}

		_, err = upsertSystem.Exec(s.GetSymbol(), data)
		if err != nil {
			panic(err)
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) GetFleet() ([]api.Ship, error) {
	bufs := [][]byte{}
	if err := r.db.Select(&bufs, "SELECT data FROM fleet"); err != nil {
		return nil, err
	}

	ships := make([]api.Ship, len(bufs))
	for i := range bufs {
		s := &api.Ship{}
		if err := s.UnmarshalJSON(bufs[i]); err != nil {
			return nil, err
		}
		ships[i] = *s
	}
	return ships, nil
}
