package models

import "github.com/bwiggs/spacetraders-go/api"

type Market struct {
	api.Market
}

func (m Market) GetShipSellableGoods(ship *Ship) []api.SellCargoReq {
	res := []api.SellCargoReq{}
	for _, imp := range m.Imports {
		for _, inv := range ship.Cargo.Inventory {
			if inv.Symbol == imp.Symbol {
				res = append(res, api.SellCargoReq{
					Symbol: inv.Symbol,
					Units:  inv.Units,
				})
			}
		}
	}

	for _, ex := range m.Exchange {
		for _, inv := range ship.Cargo.Inventory {
			if inv.Symbol == ex.Symbol {
				res = append(res, api.SellCargoReq{
					Symbol: inv.Symbol,
					Units:  inv.Units,
				})
			}
		}
	}
	return res
}
