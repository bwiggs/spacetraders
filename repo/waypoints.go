package repo

import (
	"log/slog"

	"github.com/bwiggs/spacetraders-go/api"
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repo) UpsertWaypoints(wps []api.Waypoint) error {
	log := slog.With("action", "UpsertWaypoints")
	log.Info("starting transaction")
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	log.Info("starting prepared statement")
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO waypoints (symbol, type, x, y, is_market, is_shipyard) values (?, ?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	log.Info("looping")

	for _, wp := range wps {
		log.Info("processing: " + string(wp.Symbol))
		isMarketplace := false
		isShipYard := false
		for _, t := range wp.Traits {
			switch t.Symbol {
			case "MARKETPLACE":
				isMarketplace = true
			case "SHIPYARD":
				isShipYard = true
			}
		}
		_, err = stmt.Exec(wp.Symbol, wp.Type, wp.X, wp.Y, isMarketplace, isShipYard)
		if err != nil {
			return err
		}
	}
	log.Info("commit")

	tx.Commit()

	return nil
}

func (r *Repo) GetSystemMarkets(system string) ([]string, error) {
	slog.Info(system)
	rows, err := r.db.Query("SELECT symbol FROM waypoints WHERE is_market = 1")
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

func (r *Repo) GetSystemShipyards(system string) ([]string, error) {
	slog.Info(system)
	rows, err := r.db.Query("SELECT symbol FROM waypoints WHERE is_shipyard = 1")
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
