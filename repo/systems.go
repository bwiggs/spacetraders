package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repo) UpsertSystems(systems []api.System) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertSystem, err := tx.Prepare("INSERT OR REPLACE INTO systems (name, symbol, constellation, sector_symbol, type, x, y) values (?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer upsertSystem.Close()

	for _, s := range systems {
		// upsert Market waypoints
		_, err = upsertSystem.Exec(s.GetName().Value, s.GetSymbol(), s.GetConstellation().Value, s.GetSectorSymbol(), s.GetType(), s.GetX(), s.GetY())
		if err != nil {
			panic(err)
		}
	}

	tx.Commit()

	return nil
}
