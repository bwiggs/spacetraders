package actors

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/pkg/errors"
)

type Ship struct {
	symbol  string
	state   *api.Ship
	mission Mission
	log     *slog.Logger
	client  *api.Client
}

func NewShip(ship *api.Ship, client *api.Client) *Ship {
	s := &Ship{
		symbol: ship.Symbol,
		state:  ship,
		client: client,
		log:    slog.With("ship", ship.Symbol),
	}

	go func(ship *Ship) {
		ship.log.Info("spinning up ship loop")
		for {
			if ship.mission != nil {
				ship.mission.Execute(s)
				ship.Cooldown()
			} else {
				ship.log.Debug("idling - no current mission")
				time.Sleep(1 * time.Second)
			}
			time.Sleep(3 * time.Second)
		}
	}(s)

	return s
}

func (s *Ship) SetMission(mission Mission) {
	s.log.Info("new mission " + reflect.TypeOf(mission).String())
	s.mission = mission
}

func (s *Ship) Sell(good string, wp string) error {
	s.log.Info("Selling " + good)
	if !s.IsDocked() {
		if err := s.Dock(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	for {
		ownedUnits := s.CountInventoryBySymbol(good)
		if ownedUnits == 0 {
			break
		}

		res, err := s.client.GetMarket(
			context.TODO(),
			api.GetMarketParams{SystemSymbol: wp[:7], WaypointSymbol: wp},
		)
		if err != nil {
			return errors.Wrap(err, "Selling: Failed to get market info")
		}

		var maxTradeVol int
		for _, g := range res.Data.TradeGoods {
			if g.Symbol == api.TradeSymbol(good) {
				maxTradeVol = g.TradeVolume
				break
			}
		}

		if maxTradeVol == 0 {
			s.log.Info("Selling: not accepting any units")
			break
		}

		units := min(ownedUnits, maxTradeVol)
		s.log.Info(fmt.Sprintf("Selling %d units", units))

		sres, err := s.client.SellCargo(
			context.TODO(),
			api.NewOptSellCargoReq(api.SellCargoReq{
				Symbol: api.TradeSymbol(good),
				Units:  units,
			}),
			api.SellCargoParams{ShipSymbol: s.symbol},
		)
		if err != nil {
			return errors.Wrap(err, "Sell failed")
		}

		s.state.Cargo = sres.Data.Cargo
	}

	return nil
}

func (s *Ship) Buy(good string, wp string) error {
	s.log.Info("Buying " + good)
	if !s.IsDocked() {
		if err := s.Dock(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	for {

		cargoSpace := s.state.Cargo.Capacity - s.state.Cargo.Units

		if cargoSpace == 0 {
			break
		}

		res, err := s.client.GetMarket(
			context.TODO(),
			api.GetMarketParams{SystemSymbol: wp[:7], WaypointSymbol: wp},
		)
		if err != nil {
			return errors.Wrap(err, "Buying: Failed to get market info")
		}

		var availableUnits int
		for _, g := range res.Data.TradeGoods {
			if g.Symbol == api.TradeSymbol(good) {
				availableUnits = g.TradeVolume
				break
			}
		}

		if availableUnits == 0 {
			s.log.Info("Buying: no units available for purchase")
			break
		}

		units := min(cargoSpace, availableUnits)
		s.log.Info(fmt.Sprintf("Buying %d units", units))

		pres, err := s.client.PurchaseCargo(
			context.TODO(),
			api.NewOptPurchaseCargoReq(api.PurchaseCargoReq{
				Symbol: api.TradeSymbol(good),
				Units:  units,
			}),
			api.PurchaseCargoParams{ShipSymbol: s.symbol},
		)
		if err != nil {
			return errors.Wrap(err, "Buy failed")
		}

		s.state.Cargo = pres.Data.Cargo
	}

	return nil
}

func (s *Ship) Transit(dest string) error {
	s.log.Info("Transiting " + dest)
	if s.IsDocked() {
		if err := s.Undock(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}
	navReq := api.NewOptNavigateShipReq(api.NavigateShipReq{WaypointSymbol: dest})
	res, err := s.client.NavigateShip(context.TODO(), navReq, api.NavigateShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "NavigateShip failed")
	}

	// update state
	s.state.Fuel = res.Data.Fuel
	s.state.Nav = res.Data.Nav

	travelDuration := time.Until(s.state.Nav.Route.Arrival)
	s.log.Info(fmt.Sprintf("transitting for %f seconds. sleeping", travelDuration.Seconds()))
	time.Sleep(travelDuration)
	return nil
}

func (s *Ship) Dock() error {
	s.log.Info("Docking")
	res, err := s.client.DockShip(context.TODO(), api.DockShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "Dock failed")
	}

	s.state.Nav = res.Data.Nav

	return nil
}

func (s *Ship) Undock() error {
	s.log.Info("Undocking")
	res, err := s.client.OrbitShip(context.TODO(), api.OrbitShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "Undock failed")
	}

	s.state.Nav = res.Data.Nav

	return nil
}

func (s *Ship) Refuel() error {
	s.log.Debug("Refueling")
	units := s.state.Fuel.Capacity - s.state.Fuel.Current

	if units <= 0 {
		s.log.Debug("Refueling: tanks full")
		return nil
	}

	s.log.Info(fmt.Sprintf("Refueling: %d units", units))

	if !s.IsDocked() {
		if err := s.Dock(); err != nil {
			return errors.Wrap(err, "Refuel failed to dock")
		}
	}

	req := api.RefuelShipReq{Units: api.NewOptInt(units)}
	res, err := s.client.RefuelShip(context.TODO(), api.NewOptRefuelShipReq(req), api.RefuelShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "Refuel failed")
	}

	s.state.Fuel = res.Data.Fuel

	return nil
}

func (s *Ship) At(wp string) bool {
	return s.state.Nav.WaypointSymbol == api.WaypointSymbol(wp)
}

func (s *Ship) HasGood(good string) bool {
	return s.CountInventoryBySymbol(good) > 0
}

func (s *Ship) IsCargoFull() bool {
	return s.state.Cargo.Units == s.state.Cargo.Capacity
}

func (s *Ship) IsCargoEmpty() bool {
	return s.state.Cargo.Units == 0
}

func (s *Ship) IsDocked() bool {
	return s.state.Nav.Status == api.ShipNavStatusDOCKED
}

func (s *Ship) InTransit() bool {
	return s.state.Nav.Status == api.ShipNavStatusINTRANSIT
}

func (s *Ship) Cooldown() {
	t := time.Duration(s.state.Cooldown.RemainingSeconds)
	if t > 0 {
		s.log.Info(fmt.Sprintf("cooling down: %d sec", t))
		time.Sleep(t * time.Second)
	} else {
		s.log.Debug("no cool down")
	}
}
func (s *Ship) InCooldown() bool {
	return s.state.Cooldown.RemainingSeconds > 0
}

func (s *Ship) IsFuelFull() bool {
	return s.state.Fuel.Current == s.state.Fuel.Capacity
}

func (s *Ship) CountInventoryBySymbol(tradeSymbol string) int {
	for _, inv := range s.state.Cargo.Inventory {
		if inv.Symbol == api.TradeSymbol(tradeSymbol) {
			return inv.Units
		}
	}
	return 0
}
