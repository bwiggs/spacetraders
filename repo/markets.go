package repo

import (
	"github.com/bwiggs/spacetraders-go/api"
)

func (r *Repo) UpsertMarket(market api.Market) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertMarketStmt, err := tx.Prepare("INSERT OR REPLACE INTO shipyards (waypoint, ship, type) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer upsertMarketStmt.Close()

	for _, tg := range market.Exchange {
		_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "exchange")
		if err != nil {
			return err
		}
	}

	for _, tg := range market.Exports {
		_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "exports")
		if err != nil {
			return err
		}
	}

	for _, tg := range market.Imports {
		_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "imports")
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}
