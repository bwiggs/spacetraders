package repo

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/utils"
	"github.com/pkg/errors"
)

func (r *Repo) UpsertMarket(market api.Market) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	err = r.UpsertTradeGoods(market.Exchange)
	if err != nil {
		return errors.Wrap(err, "UpsertTradeGoods exchange")
	}
	err = r.UpsertTradeGoods(market.Exports)
	if err != nil {
		return errors.Wrap(err, "UpsertTradeGoods exports")
	}
	err = r.UpsertTradeGoods(market.Imports)
	if err != nil {
		return errors.Wrap(err, "UpsertTradeGoods imports")
	}

	if market.TradeGoods != nil {
		upsertMarketStmt, err := tx.Prepare("INSERT OR REPLACE INTO markets (waypoint, good, type, volume, activity, bid, ask) values (?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		defer upsertMarketStmt.Close()

		for _, tg := range market.TradeGoods {
			_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, tg.Type, tg.TradeVolume, tg.Activity.Value, tg.PurchasePrice, tg.SellPrice)
			if err != nil {
				return err
			}
		}
	} else {
		upsertMarketStmt, err := tx.Prepare("INSERT OR REPLACE INTO markets (waypoint, good, type) values (?, ?, ?)")
		if err != nil {
			return err
		}
		defer upsertMarketStmt.Close()

		for _, tg := range market.Exchange {
			_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "EXCHANGE")
			if err != nil {
				return err
			}
		}

		for _, tg := range market.Exports {
			_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "EXPORT")
			if err != nil {
				return err
			}
		}

		for _, tg := range market.Imports {
			_, err = upsertMarketStmt.Exec(market.Symbol, tg.Symbol, "IMPORT")
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func (r *Repo) UpsertTradeGoods(goods []api.TradeGood) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	upsert, err := tx.Prepare("INSERT OR REPLACE INTO goods (symbol, name, description) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer upsert.Close()

	for _, good := range goods {
		_, err = upsert.Exec(good.Symbol, good.Name, good.Description)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

// FindMarketsForGoods returns a list of markets which the provided goods could
// be sold at. Markets are returned with the top result accepting the most
// goods, and the bottom accepting the least goods.
func (r *Repo) FindMarketsForGoods(goods []string) ([]string, error) {

	params := strings.Join(strings.Split(strings.Repeat("?", len(goods)), ""), ", ")
	sql := fmt.Sprintf(`WITH results AS (
	SELECT 
		waypoint, 
		count(*) as sellables 
	FROM markets 
	WHERE 
		type = 'IMPORT' AND 
		good IN (%s) 
	GROUP BY waypoint 
	ORDER BY sellables DESC) 
SELECT waypoint FROM results`, params)

	// this is dumb
	s := make([]any, len(goods))
	for i, v := range goods {
		s[i] = v
	}

	rows, err := r.db.Query(sql, s...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	waypoints := make([]string, 0)
	for rows.Next() {
		var wp string
		if err := rows.Scan(&wp); err != nil {
			log.Fatal(err)
		}
		waypoints = append(waypoints, wp)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return waypoints, nil
}

// FindMarketsWithGoods returns a list of markets which the provided goods could
// be sold at. Markets are returned with the top result accepting the most
// goods, and the bottom accepting the least goods.
func (r *Repo) FindMarketsWithGoods(goods []string) ([]string, error) {

	params := strings.Join(strings.Split(strings.Repeat("?", len(goods)), ""), ", ")
	sql := fmt.Sprintf(`with results as (SELECT waypoint, count(*) as sellables from markets where type = 'IMPORT' and good in (%s) group by waypoint order by sellables desc) select waypoint from results`, params)

	// this is dumb
	s := make([]any, len(goods))
	for i, v := range goods {
		s[i] = v
	}

	rows, err := r.db.Query(sql, s...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	waypoints := make([]string, 0)
	for rows.Next() {
		var wp string
		if err := rows.Scan(&wp); err != nil {
			log.Fatal(err)
		}
		waypoints = append(waypoints, wp)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return waypoints, nil
}

// FindExportWaypointsForGood returns a list of markets which are exporting the provided good ordered by ask price asc.
func (r *Repo) FindExportWaypointsForGood(good string) ([]string, error) {
	rows, err := r.db.Query(`select waypoint from markets where good = ? and type = 'EXPORT' order by ask asc`, good)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	waypoints := make([]string, 0)
	for rows.Next() {
		var wp string
		if err := rows.Scan(&wp); err != nil {
			log.Fatal(err)
		}
		waypoints = append(waypoints, wp)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return waypoints, nil
}

type MarketTrade struct {
	Good  string `sql:"good"`
	Gross int    `sql:"gross"`

	Origin  string `sql:"origin"`
	OriginX int    `sql:"ox"`
	OriginY int    `sql:"oy"`
	Bid     int    `sql:"bid"`

	Dest  string `sql:"dest"`
	DestX int    `sql:"dx"`
	DestY int    `sql:"dy"`
	Ask   int    `sql:"bid"`

	Distance int
}

// FindMarketTrades
func (r *Repo) FindMarketTrades() ([]MarketTrade, error) {
	rows, err := r.db.Query(`SELECT 
	im.good,
	ex.waypoint AS origin,
	im.ask - ex.bid AS gross,	
	ex.bid,
	wex.x AS ox,
	wex.y AS oy,
	im.waypoint AS dest,
	im.ask,
	wim.x AS dx,
	wim.y AS dy
FROM markets im
JOIN markets ex ON im.good = ex.good AND im.type = 'IMPORT' AND ex.type = 'EXPORT' AND im.bid IS NOT NULL AND ex.ask IS NOT NULL
JOIN waypoints wim ON wim.symbol = im.waypoint
JOIN waypoints wex ON wex.symbol = ex.waypoint
WHERE
	gross > 100 AND
	abs(wex.x) <= 100 AND abs(wex.y) <= 100 AND abs(wim.x) <= 100 AND abs(wim.y) <= 100
ORDER BY gross DESC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trades := []MarketTrade{}
	for rows.Next() {
		mt := MarketTrade{}
		if err := rows.Scan(
			&mt.Good,
			&mt.Origin,
			&mt.Gross,
			&mt.Bid,
			&mt.OriginX,
			&mt.OriginY,
			&mt.Dest,
			&mt.Ask,
			&mt.DestX,
			&mt.DestY,
		); err != nil {
			log.Fatal(err)
		}

		mt.Distance = utils.Distance2dInt(mt.OriginX, mt.OriginY, mt.DestX, mt.DestY)
		trades = append(trades, mt)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return trades, nil
}
