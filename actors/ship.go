package actors

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/pkg/errors"
)

type Ship struct {
	state        *api.Ship
	symbol       string
	mission      Mission
	log          *slog.Logger
	client       *api.Client
	transferLock sync.Mutex
}

func NewShip(ship *api.Ship, client *api.Client) *Ship {

	logger := slog.With("ship", ship.Symbol)

	s := &Ship{
		symbol:       ship.Symbol,
		state:        ship,
		client:       client,
		log:          logger,
		transferLock: sync.Mutex{},
	}

	go func(ship *Ship) {
		data := Blackboard{ship: ship, log: logger}
		for {
			ship.Wait()
			if ship.mission != nil {
				ship.mission.Execute(&data)
			} else {
				ship.log.Info("idling - no current mission")
			}
		}
	}(s)

	return s
}

func (s *Ship) Wait() {
	var dur time.Duration

	arrivalTime := time.Until(s.state.Nav.Route.Arrival)
	cooldownTime := time.Until(s.state.Cooldown.Expiration.Value)

	if arrivalTime > 0 {
		dest := s.state.Nav.Route.Destination.Symbol
		dur = arrivalTime
		s.log.Info(fmt.Sprintf("%s: transit: %s %s", s.symbol, dest, dur), "route.dest", dest)
	} else if cooldownTime > 0 {
		dur = cooldownTime
		s.log.Info(fmt.Sprintf("%s: cooldown: %s", s.symbol, dur))
	} else {
		dur = 2 * time.Second
	}

	time.Sleep(dur)
}

func (s *Ship) SetMission(mission Mission) {
	s.log.Info("Mission: " + mission.String())
	s.mission = mission
}

func (s *Ship) DeliverContract(contractID, good string) (*api.Contract, error) {
	s.log.Info("DeliverContract: " + good)
	if !s.IsDocked() {
		if err := s.Dock(); err != nil {
			return nil, errors.Wrap(err, "Undock failed")
		}
	}

	ownedUnits := s.CountInventoryBySymbol(good)
	s.log.Debug(fmt.Sprintf("ship has %d units of %s", ownedUnits, good))
	if ownedUnits == 0 {
		s.log.Debug("no units to deliver")
		return nil, nil
	}

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

func (s *Ship) SellCargo() error {
	res, err := s.client.GetMarket(
		context.TODO(),
		api.GetMarketParams{SystemSymbol: s.CurrWaypoint()[:7], WaypointSymbol: s.CurrWaypoint()},
	)
	if err != nil {
		return errors.Wrap(err, "Selling: Failed to get market info")
	}

	if r, err := repo.GetRepo(); err != nil {
		if err := r.UpsertMarket(res.Data); err != nil {
			s.log.Warn(errors.Wrap(err, "failed to update market repo data").Error())
		}
	}

	marketItems := make(map[string]int)
	for _, i := range res.Data.TradeGoods {
		if i.TradeVolume == 0 {
			continue
		}
		marketItems[string(i.Symbol)] = i.TradeVolume
	}

	for _, inv := range s.state.Cargo.Inventory {
		good := string(inv.Symbol)
		vol, hasItem := marketItems[good]
		if !hasItem {
			continue
		}

		units := min(inv.Units, vol)
		s.log.Info(fmt.Sprintf("Selling %d %s", units, good))

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

		t := sres.Data.Transaction

		s.log.Info(fmt.Sprintf("Transaction: Sell: $(%+d) %s (%d x $%d)", t.TotalPrice, t.TradeSymbol, t.Units, t.PricePerUnit))
		s.state.Cargo = sres.Data.Cargo
	}

	return nil
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

func (s *Ship) Buy(good string, maxUnits int, wp string) error {
	s.log.Info("Buying " + good)

	for {

		cargoSpace := s.state.Cargo.Capacity - s.state.Cargo.Units

		if cargoSpace == 0 {
			s.log.Debug("Cargo full, cant buy any more goods")
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

		units := min(cargoSpace, availableUnits, maxUnits)
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

func (s *Ship) GetCargoItemBySymbol(good api.TradeSymbol) (api.ShipCargoItem, bool) {
	for _, inv := range s.state.Cargo.Inventory {
		if inv.Symbol == good {
			return inv, true
		}
	}
	return api.ShipCargoItem{}, false
}

func (s *Ship) ReceiveTransfer(from *Ship, tradeGoodSymbol api.TradeSymbol, maxUnits int) (bool, error) {
	if s.IsCargoFull() {
		s.log.Info("cargo full")
		return false, nil
	}

	item, found := from.GetCargoItemBySymbol(tradeGoodSymbol)
	if !found {
		return false, nil
	}

	s.transferLock.Lock()
	defer s.transferLock.Unlock()

	units := min(item.Units, s.AvailableCargoUnits())
	if maxUnits > 0 {
		units = min(units, maxUnits)
	}

	res, err := s.client.TransferCargo(
		context.TODO(),
		api.NewOptTransferCargoReq(api.TransferCargoReq{
			TradeSymbol: item.Symbol,
			Units:       units,
			ShipSymbol:  s.symbol,
		}),
		api.TransferCargoParams{ShipSymbol: from.symbol},
	)
	if err != nil {
		return false, err
	}

	from.state.Cargo = res.Data.Cargo
	s.state.Cargo.Units += units

	s.log.Info("transfer received", "cargo.units", s.state.Cargo.Units, "transfer.from", from.symbol, "transfer.units", units, "transfer.good", item.Symbol)

	return true, nil
}

func (s *Ship) Transit(dest string) error {
	if s.CurrWaypoint() == dest {
		return nil
	}

	if s.IsDocked() {
		if err := s.Orbit(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	navReq := api.NewOptNavigateShipReq(api.NavigateShipReq{WaypointSymbol: dest})
	res, err := s.client.NavigateShip(context.TODO(), navReq, api.NavigateShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "NavigateShip failed: maybe not enough fuel?")
	}

	// update state
	s.state.Fuel = res.Data.Fuel
	s.state.Nav = res.Data.Nav

	dur := time.Until(s.state.Nav.Route.Arrival)
	s.log.Info("transiting", "nav.route.dest", dest, "nav.route.arrival", dur)
	return nil
}

func (s *Ship) Survey() ([]api.Survey, error) {
	if !s.HasSurveyor() {
		return nil, nil
	}

	if s.IsDocked() {
		if err := s.Orbit(); err != nil {
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

func (s *Ship) Orbit() error {
	s.log.Info("Orbiting")
	res, err := s.client.OrbitShip(context.TODO(), api.OrbitShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return errors.Wrap(err, "Orbit failed")
	}

	s.state.Nav = res.Data.Nav

	return nil
}

func (s *Ship) Refuel() error {
	units := s.state.Fuel.Capacity - s.state.Fuel.Current

	if units <= 0 {
		return nil
	}

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
	s.log.Info(fmt.Sprintf("Refueled: $%d (%d x $%d)", res.Data.Transaction.TotalPrice, units, res.Data.Transaction.PricePerUnit))

	return nil
}

func (s *Ship) AvailableCargoUnits() int {
	return s.state.Cargo.Capacity - s.state.Cargo.Units
}

func (s *Ship) HasFreeCargoSpace() bool {
	return s.AvailableCargoUnits() > 0
}

func (s *Ship) At(wp string) bool {
	return s.state.Nav.WaypointSymbol == api.WaypointSymbol(wp)
}

func (s *Ship) Jettison(good string) error {
	if s.IsDocked() {
		if err := s.Orbit(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
	}

	units := s.CountInventoryBySymbol(good)

	if units <= 0 {
		return nil
	}

	l := s.log.With("good", good, "units", units)

	res, err := s.client.Jettison(
		context.TODO(),
		api.NewOptJettisonReq(api.JettisonReq{Symbol: api.TradeSymbol(good), Units: units}),
		api.JettisonParams{ShipSymbol: s.symbol},
	)
	if err != nil {
		return errors.Wrap(err, "jettison failed:")
	}

	l.Info("jettisoned")

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
	return s.state.Cargo.Units >= s.state.Cargo.Capacity
}

func (s *Ship) IsCargoEmpty() bool {
	return s.state.Cargo.Units == 0
}

func (s *Ship) IsDocked() bool {
	return s.state.Nav.Status == api.ShipNavStatusDOCKED
}

func (s *Ship) IsIdle() bool {
	return !(s.InTransit() || s.InCooldown())
}

func (s *Ship) InTransit() bool {
	return s.state.Nav.Route.Arrival.After(time.Now())
}

func (s *Ship) Cooldown() {
	t := time.Duration(s.state.Cooldown.RemainingSeconds)
	if t > 0 {
		s.log.Info(fmt.Sprintf("Cooldown: sleeping: %s", t))
		time.Sleep(t * time.Second)
	}
}

func (s *Ship) Extract(survey api.Survey) error {
	if s.IsCargoFull() {
		return nil
	}

	if s.IsDocked() {
		if err := s.Orbit(); err != nil {
			return errors.Wrap(err, "Undock failed")
		}
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

	s.log.Info(fmt.Sprintf("%s: Extracted %d %s", s.state.Symbol, yield.Units, yield.Symbol), "cargo.units", s.state.Cargo.Units, "cargo.cap", s.state.Cargo.Capacity)

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

func (s *Ship) Update() error {
	res, err := s.client.GetMyShip(context.TODO(), api.GetMyShipParams{ShipSymbol: s.symbol})
	if err != nil {
		return err
	}
	s.state = &res.Data
	return nil
}
