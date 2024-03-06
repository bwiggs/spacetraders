package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repo) UpsertWaypoints(wps []api.Waypoint) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertWaypoint, err := tx.Prepare("INSERT OR REPLACE INTO waypoints (symbol, type, x, y) values (?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer upsertWaypoint.Close()

	upsertTrait, err := tx.Prepare("INSERT OR REPLACE INTO traits (symbol, name, description) values (?, ?, ?);")
	if err != nil {
		return err
	}
	defer upsertTrait.Close()

	upsertWaypointTrait, err := tx.Prepare("INSERT OR REPLACE INTO waypoints_traits (waypoint, trait) values (?, ?);")
	if err != nil {
		return err
	}
	defer upsertWaypointTrait.Close()

	for _, wp := range wps {
		// upsert Market waypoints
		_, err = upsertWaypoint.Exec(wp.Symbol, wp.Type, wp.X, wp.Y)
		if err != nil {
			return err
		}

		for _, t := range wp.Traits {
			_, err = upsertTrait.Exec(t.Symbol, t.Name, t.Description)
			if err != nil {
				return err
			}

			_, err = upsertWaypointTrait.Exec(wp.Symbol, t.Symbol)
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) GetSystemWaypointsByTrait(system, trait string) ([]string, error) {
	rows, err := r.db.Query("select distinct waypoint from waypoints_traits wt where wt.trait = ?", trait)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	symbols := []string{}
	for rows.Next() {
		var symbol string
		err = rows.Scan(&symbol)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return symbols, nil
}
