package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func (r *Repo) UpsertWaypoints(wps []api.Waypoint) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertWaypoint, err := tx.Prepare("INSERT OR REPLACE INTO waypoints (symbol, type, x, y, json) values (?, ?, ?, ?, ?);")
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
		json, err := wp.MarshalJSON()
		if err != nil {
			return err
		}
		_, err = upsertWaypoint.Exec(wp.Symbol, wp.Type, wp.X, wp.Y, json)
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

func (r *Repo) WaypointHasTrait(waypointSymbol, trait string) (bool, error) {
	sql := `SELECT 1 FROM waypoints_traits wt WHERE wt.waypoint = ? AND wt.trait = ?`

	rows, err := r.db.Query(sql, waypointSymbol, trait)
	if err != nil {
		return false, errors.Wrap(err, "WaypointHasTrait: query failed")
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (r *Repo) GetSystemWaypointsByTrait(system, trait string) ([]*models.Waypoint, error) {
	sql := `select json 
	from waypoints 
	join waypoints_traits wt on wt.waypoint = waypoints.symbol 
	where wt.trait = ? AND waypoints.symbol LIKE ?`

	rows, err := r.db.Query(sql, trait, system+"%")
	if err != nil {
		return nil, errors.Wrap(err, "GetSystemWaypointsByTrait: failed to query waypoints")
	}
	defer rows.Close()

	waypoints := []*models.Waypoint{}

	for rows.Next() {
		buf := []byte{}

		err = rows.Scan(&buf)
		if err != nil {
			return nil, errors.Wrap(err, "GetSystemWaypointsByTrait: failed to scan sql response")
		}

		wp := api.Waypoint{}
		if err := wp.UnmarshalJSON(buf); err != nil {
			return nil, errors.Wrap(err, "GetSystemWaypointsByTrait: failed to unmarshal waypoint")
		}
		waypoints = append(waypoints, models.NewWaypoint(&wp))
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "GetSystemWaypointsByTrait: row err:")
	}

	return waypoints, nil
}

func (r *Repo) GetWaypoints(system string) ([]*models.Waypoint, error) {
	waypoints := []*models.Waypoint{}

	waypointjson := [][]byte{}
	if err := r.db.Select(&waypointjson, `SELECT json FROM waypoints`); err != nil {
		return nil, err
	}
	for _, j := range waypointjson {
		var wp api.Waypoint
		if err := wp.UnmarshalJSON(j); err != nil {
			return nil, err
		}
		waypoints = append(waypoints, models.NewWaypoint(&wp))
	}

	return waypoints, nil
}

func (r *Repo) GetNonOrbitalWaypoints(system string) ([]*models.Waypoint, error) {
	wps, err := r.GetWaypoints(system)
	if err != nil {
		return nil, err
	}

	waypoints := []*models.Waypoint{}
	for _, wp := range wps {
		if wp.Orbits.Value == "" {
			waypoints = append(waypoints, wp)
		}
	}

	return waypoints, nil
}
