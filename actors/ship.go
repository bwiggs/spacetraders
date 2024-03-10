package actors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/pkg/errors"
)

type Ship struct {
	state   *api.Ship
	symbol  string
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
		ship.log.Debug("spinning up ship loop")
		for {
			if ship.mission != nil {
				ship.mission.Execute(s)
				ship.Cooldown()
			} else {
				ship.log.Info("idling - no current mission")
			}
			time.Sleep(3 * time.Second)
		}
	}(s)

	return s
}

func (s *Ship) SetMission(mission Mission) {
	s.log.Info("Mission: " + mission.String())
	s.mission = mission
}

func (s *Ship) Deliver(contractID, good string) (*api.Contract, error) {
	s.log.Info("Delivering " + good)
	if !s.IsDocked() {
		if err := s.Dock(); err != nil {
			return nil, errors.Wrap(err, "Undock failed")
		}
	}

	ownedUnits := s.CountInventoryBySymbol(good)

	dcr := api.DeliverContractReq{ShipSymbol: s.symbol, TradeSymbol: good, Units: ownedUnits}

	sres, err := s.client.DeliverContract(
		context.TODO(),
		api.NewOptDeliverContractReq(dcr),
		api.DeliverContractParams{ContractId: contractID},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Contract Deliver failed")
	}

	s.state.Cargo = sres.Data.Cargo

	return &sres.Data.Contract, nil
}

func (s *Ship) Sell(good string) error {
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
			api.GetMarketParams{SystemSymbol: s.CurrWaypoint()[:7], WaypointSymbol: s.CurrWaypoint()},
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
			s.log.Info("Selling: not accepting any more units")
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
	s.log.Debug("Transiting " + dest)

	if s.CurrWaypoint() == dest {
		return nil
	}

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
	s.log.Info(fmt.Sprintf("transitting to %s - %f seconds. sleeping", dest, travelDuration.Seconds()), "dest", dest)
	time.Sleep(travelDuration)
	return nil
}

func (s *Ship) Survey() ([]api.Survey, error) {
	if !s.HasSurveyor() {
		return nil, nil
	}

	if s.IsDocked() {
		if err := s.Undock(); err != nil {
			return nil, errors.Wrap(err, "Undock failed")
		}
	}

	s.log.Info("Surveying")

	res, err := s.client.CreateSurvey(context.TODO(), api.CreateSurveyParams{ShipSymbol: s.symbol})
	if err != nil {
		return nil, errors.Wrap(err, "failed creating survey")
	}

	s.state.Cooldown = res.Data.Cooldown
	return res.Data.Surveys, nil
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

func (s *Ship) Jettison(good string) error {
	if s.IsDocked() {
		if err := s.Undock(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	units := s.CountInventoryBySymbol(good)

	s.log.Info(fmt.Sprintf("Jettison: %d %s", units, good))

	res, err := s.client.Jettison(
		context.TODO(),
		api.NewOptJettisonReq(api.JettisonReq{Symbol: api.TradeSymbol(good), Units: units}),
		api.JettisonParams{ShipSymbol: s.symbol},
	)
	if err != nil {
		return err
	}

	s.state.Cargo = res.Data.Cargo

	return nil
}

func (s *Ship) HasGood(good string) bool {
	return s.CountInventoryBySymbol(good) > 0
}

func (s *Ship) HasSurveyor() bool {
	for _, m := range s.state.Mounts {
		if m.Symbol == api.ShipMountSymbolMOUNTSURVEYORI || m.Symbol == api.ShipMountSymbolMOUNTSURVEYORII || m.Symbol == api.ShipMountSymbolMOUNTSURVEYORI {
			return true
		}
	}
	return false
}

func (s *Ship) CurrWaypoint() string {
	return string(s.state.Nav.WaypointSymbol)
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
	}
}

func (s *Ship) Extract(survey api.Survey) error {
	s.log.Info("Extracting")

	if s.IsDocked() {
		if err := s.Undock(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	for {
		if s.IsCargoFull() {
			break
		}

		var yield api.ExtractionYield

		if time.Until(survey.Expiration) > 0 {
			s.log.Info("leveraging survey")

			os := api.NewOptSurvey(survey)
			res, err := s.client.ExtractResourcesWithSurvey(context.TODO(), os, api.ExtractResourcesWithSurveyParams{ShipSymbol: s.symbol})
			if err != nil {
				return errors.Wrap(err, "ExtractResources failed")
			}
			s.state.Cooldown = res.Data.Cooldown
			s.state.Cargo = res.Data.Cargo
			yield = res.Data.Extraction.Yield
		} else {
			res, err := s.client.ExtractResources(context.TODO(), api.OptExtractResourcesReq{}, api.ExtractResourcesParams{ShipSymbol: s.symbol})
			if err != nil {
				return errors.Wrap(err, "ExtractResources failed")
			}
			s.state.Cooldown = res.Data.Cooldown
			s.state.Cargo = res.Data.Cargo
			yield = res.Data.Extraction.Yield
		}

		s.log.Info(fmt.Sprintf("Extracted %d %s", yield.Units, yield.Symbol))
		s.log.Info(fmt.Sprintf("Cooling down for %s", time.Until(s.state.Cooldown.Expiration.Value)))

		time.Sleep(time.Until(s.state.Cooldown.Expiration.Value))

	}

	return nil
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

func (s *Ship) InventorySymbols() []string {
	invs := make([]string, len(s.state.Cargo.Inventory))
	for i, inv := range s.state.Cargo.Inventory {
		invs[i] = string(inv.Symbol)
	}
	return invs
}
