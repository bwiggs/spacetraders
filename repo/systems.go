package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func (r *Repo) UpsertSystems(systems []api.System) error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err, "UpsertSystems: failed to start tx")
	}
	upsertSystem, err := tx.Prepare("INSERT OR REPLACE INTO systems (name, symbol, constellation, sector_symbol, type, x, y) values (?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		return errors.Wrap(err, "UpsertSystems: failed to prepare query")
	}
	defer upsertSystem.Close()

	for _, s := range systems {
		// upsert Market waypoints
		_, err = upsertSystem.Exec(s.GetName().Value, s.GetSymbol(), s.GetConstellation().Value, s.GetSectorSymbol(), s.GetType(), s.GetX(), s.GetY())
		if err != nil {
			return errors.Wrap(err, "UpsertSystems: fail")
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) GetSystems() ([]models.System, error) {
	systems := []models.System{}
	err := r.db.Select(&systems, "SELECT symbol, constellation, name, type, sector_symbol, x, y FROM systems")
	if err != nil {
		return nil, errors.Wrap(err, "GetSystems: failed to select systems")
	}
	return systems, nil
}
