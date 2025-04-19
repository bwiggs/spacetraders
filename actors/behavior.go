package actors

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/go-faster/errors"
)

func NavigationAction() *bt.Selector {
	return bt.NewSelector(
		ConditionIsAtNavDest{},
		CheckInRouteToDestination{},
		bt.NewSequence(
			RefuelAction{},
			OrbitAction{},
			NavAction{},
		),
	)
}

type CheckInRouteToDestination struct{}

func (a CheckInRouteToDestination) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("CheckInRouteToDestination: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("CheckInRouteToDestination: blackboard: ship was nil")
		return bt.Running
	}

	if bb.ship.state.Nav.Status != api.ShipNavStatusINTRANSIT {
		return bt.Failure
	}

	if bb.destination == string(bb.ship.state.Nav.Route.Destination.Symbol) {
		return bt.Running
	}

	return bt.Failure
}

type ConditionContractIsActive struct{}

func (a ConditionContractIsActive) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionContractIsActive: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionContractIsActive: blackboard: contract was nil")
		return bt.Running
	}

	if bb.contract.Fulfilled {
		return bt.Failure
	}

	if time.Until(bb.contract.Terms.Deadline) > 0 {
		return bt.Success
	}

	return bt.Failure
}

type ConditionIsProfitableTradeRouteForContract struct{}

func (a ConditionIsProfitableTradeRouteForContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionIsProfitableTradeRouteForContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		bb.Logger().Error("ConditionIsProfitableTradeRouteForContract: blackboard: contract was nil")
		return bt.Running
	}

	// rev := bb.contract.Revenue()
	// trips :=
	// travelCost :=
	bb.Logger().Warn("ConditionIsProfitableTradeRouteForContract: unprofitable trade setup for contract")

	return bt.Failure
}

type ConditionShipHasRemainingContractUnits struct{}

func (a ConditionShipHasRemainingContractUnits) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionShipHasRemainingContractUnits: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionShipHasRemainingContractUnits: blackboard: contract was nil")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("ConditionShipHasRemainingContractUnits: blackboard: ship was nil")
		return bt.Running
	}

	for _, g := range bb.contract.Terms.Deliver {
		remaining := g.UnitsRequired - g.UnitsFulfilled
		if remaining == 0 {
			continue
		}
		for _, inv := range bb.ship.state.Cargo.Inventory {
			if string(inv.Symbol) != g.TradeSymbol {
				continue
			}
			if remaining == inv.Units {
				return bt.Success
			}
		}
	}

	return bt.Failure
}

type ConditionContractIsFulfilled struct{}

func (a ConditionContractIsFulfilled) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionContractIsFulfilled: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionContractIsFulfilled: blackboard: contract was nil")
		return bt.Running
	}

	if bb.contract.Fulfilled {
		return bt.Success
	}

	return bt.Failure
}

type ConditionIsAtExtractionWaypoint struct{}

func (a ConditionIsAtExtractionWaypoint) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionIsAtExtractionWaypoint: expected a blackboard")
		return bt.Running
	}

	if bb.extractionWaypoint == "" {
		slog.Error("ConditionIsAtExtractionWaypoint: blackboard: contract was blank")
		return bt.Running
	}

	if bb.ship.CurrWaypoint() == bb.extractionWaypoint {
		if time.Until(bb.ship.state.Nav.Route.Arrival) > 0 {
			return bt.Running
		}
		return bt.Success
	}
	return bt.Failure
}

type SetDestinationToExtractionWaypoint struct{}

func (a SetDestinationToExtractionWaypoint) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("SetDestinationToExtractionWaypoint: expected a blackboard")
		return bt.Running
	}

	if bb.extractionWaypoint == "" {
		slog.Error("SetDestinationToExtractionWaypoint: blackboard: contract was blank")
		return bt.Running
	}

	bb.destination = bb.extractionWaypoint
	return bt.Success
}

type SetDestinationToBestTradeMarketSale struct{}

func (a SetDestinationToBestTradeMarketSale) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("SetDestinationToBestTradeMarketSale: expected a blackboard")
		return bt.Running
	}

	l := bb.Logger().With("behavior", "SetDestinationToBestTradeMarketSale")

	if bb.ship == nil {
		l.Error("SetDestinationToBestTradeMarketSale: blackboard: ship was nil")
		return bt.Running
	}

	bb.ship.Update()

	trades, err := bb.repo.FindMarketTrades()
	if err != nil {
		l.Error(err.Error())
		return bt.Running
	}

	if len(trades) == 0 {
		l.Warn("no markets for trades")
		return bt.Failure
	}

	l.Info(fmt.Sprintf("SetDestinationToBestTradeMarket: options: %v", trades))
	l.Info("SetDestinationToBestTradeMarket: " + trades[0].Origin)
	bb.destination = trades[0].Origin
	bb.purchaseTargetGood = trades[0].Good
	bb.purchaseMaxUnits = 20

	return bt.Success
}

type SetDestinationToBestTradeMarket struct{}

func (a SetDestinationToBestTradeMarket) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("SetDestinationToBestTradeMarket: expected a blackboard")
		return bt.Running
	}

	l := bb.Logger().With("behavior", "SetDestinationToBestTradeMarket")

	if bb.ship == nil {
		l.Error("SetDestinationToBestTradeMarket: blackboard: ship was nil")
		return bt.Running
	}

	bb.ship.Update()

	trades, err := bb.repo.FindMarketTrades()
	if err != nil {
		l.Error(err.Error())
		return bt.Running
	}

	if len(trades) == 0 {
		l.Warn("no markets for trades")
		return bt.Failure
	}

	l.Info(fmt.Sprintf("SetDestinationToBestTradeMarket: options: %v", trades))
	l.Info("SetDestinationToBestTradeMarket: " + trades[0].Origin)
	bb.destination = trades[0].Origin
	bb.purchaseTargetGood = trades[0].Good
	bb.purchaseMaxUnits = 20

	return bt.Success
}

type SetDestinationToBestMarketToSellCargo struct{}

func (a SetDestinationToBestMarketToSellCargo) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("SetDestinationToBestMarketToSellCargo: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		bb.Logger().Error("SetDestinationToBestMarketToSellCargo: blackboard: ship was nil")
		return bt.Running
	}

	bb.ship.Update()

	markets, err := bb.repo.FindMarketsForGoods(bb.ship.InventorySymbols())
	if err != nil {
		bb.Logger().Error(errors.Wrap(err, "SetDestinationToBestMarketToSellCargo failed:").Error())
		return bt.Running
	}

	if len(markets) == 0 {
		bb.Logger().Warn(fmt.Sprintf("no markets for remaining goods: %s", bb.ship.InventorySymbols()))
		return bt.Failure
	}

	bb.destination = markets[0]

	return bt.Success
}

type SetPurchaseFromContract struct{}

func (a SetPurchaseFromContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("SetPurchaseFromContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("SetPurchaseFromContract: blackboard: contract was nil")
		return bt.Running
	}

	for _, d := range bb.contract.Terms.Deliver {
		remaining := d.UnitsRequired - d.UnitsFulfilled
		if d.UnitsRequired-d.UnitsFulfilled == 0 {
			continue
		}

		slog.Info("SetPurchaseFromContract: setting good to " + d.TradeSymbol)
		bb.purchaseTargetGood = d.TradeSymbol
		bb.purchaseMaxUnits = remaining

		// SET THE WAYPOINT TO BUY THE GOOD FROM

		markets, err := bb.repo.FindExportWaypointsForGood(d.TradeSymbol)
		if err != nil {
			slog.Error(errors.Wrap(err, "SetPurchaseFromContract: FindMarketsWithGoods failed:").Error())
			return bt.Failure
		}

		if len(markets) == 0 {
			slog.Warn("no markets to purchase " + d.TradeSymbol + "from")
			return bt.Failure
		}

		wp := markets[0]

		slog.Info("SetPurchaseFromContract: setting waypoint to " + wp)

		bb.destination = wp

		return bt.Success
	}

	return bt.Failure
}

type ConditionIsAtNavDest struct{}

func (a ConditionIsAtNavDest) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionIsAtNavDest: expected a blackboard")
		return bt.Running
	}

	if bb.destination == "" {
		slog.Error("ConditionIsAtNavDest: blackboard: destination was empty")
		return bt.Running
	}

	if bb.destination == bb.ship.state.Nav.Route.Destination.Symbol {
		return bt.Success
	}

	return bt.Failure
}

type SetDeliveryDestFromContract struct{}

func (a SetDeliveryDestFromContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("SetDeliveryDestFromContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("SetDeliveryDestFromContract: blackboard: contract was nil")
		return bt.Running
	}

	for _, d := range bb.contract.Terms.Deliver {
		if d.UnitsRequired-d.UnitsFulfilled == 0 {
			continue
		}
		bb.destination = d.DestinationSymbol
		return bt.Success
	}

	slog.Error("SetDeliveryDestFromContract: blackboard: no remaining contract goods to deliver")
	return bt.Failure
}

type NavAction struct{}

func (a NavAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("NavAction: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("NavAction: blackboard: ship was nil")
		return bt.Running
	}

	if bb.destination == "" {
		slog.Error("NavAction: blackboard: destination was empty")
		return bt.Running
	}

	if !bb.ship.At(bb.destination) {
		if err := bb.ship.Transit(bb.destination); err != nil {
			bb.ship.log.Error(errors.Wrap(err, "Failed to transit ship").Error())
			return bt.Failure
		}
	}

	if time.Until(bb.ship.state.Nav.Route.Arrival) > 0 {
		return bt.Running
	}

	return bt.Success
}

type PurchaseAction struct {
	dest string
}

func NewPurchaseAction(dest string) *PurchaseAction {
	return &PurchaseAction{dest}
}

func (a *PurchaseAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	slog.Debug("PurchaseAction")
	return bt.Running
}

type PurchaseContractGoodsAction struct {
}

func (a *PurchaseContractGoodsAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("PurchaseContractGoodsAction: expected a blackboard")
		return bt.Failure
	}

	contract := bb.contract
	if contract == nil {
		slog.Error("PurchaseContractGoodsAction: blackboard: contract was nil")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("PurchaseContractGoodsAction: blackboard: ship was nil")
		return bt.Running
	}

	for _, d := range contract.Terms.Deliver {
		remainingUnits := d.UnitsRequired - d.UnitsFulfilled
		bb.ship.log.Debug(fmt.Sprintf("Contract: %s %d remaining", d.TradeSymbol, remainingUnits), "action", "PurchaseContractGoodsAction")

		// if the ship is carrying the last bit of units for the current
		// contract good, then deliver them to the destination
		if remainingUnits == bb.ship.CountInventoryBySymbol(d.TradeSymbol) {
			bb.ship.log.Debug("ship is carrying remaining item, time to deliver.", "action", "PurchaseContractGoodsAction")
			return bt.Success
		}

		if remainingUnits > 0 {
			bb.ship.log.Debug("go buy remaining units", "action", "PurchaseContractGoodsAction")
			// return NewPurchaseGoodSequence(ship, d.TradeSymbol, remainingUnits, d.DestinationSymbol).Tick(data)
		}
	}

	return bt.Success
}

type RefuelAction struct{}

func (a RefuelAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("RefuelAction: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.Refuel(); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to refuel ship").Error())
		return bt.Running
	}

	return bt.Success
}

type OrbitAction struct{}

func (a OrbitAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("OrbitAction: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.Orbit(); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to orbit ship").Error())
		return bt.Running
	}

	return bt.Success
}

type DockAction struct{}

func (a DockAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("DockAction: expected a blackboard")
		return bt.Running
	}

	if !bb.ship.IsDocked() {
		if err := bb.ship.Dock(); err != nil {
			bb.ship.log.Error(errors.Wrap(err, "Failed to dock ship").Error())
			return bt.Running
		}
	}

	return bt.Success
}

type TransferCargoToNearbyTransport struct{}

func (a TransferCargoToNearbyTransport) Tick(data bt.Blackboard) bt.BehaviorStatus {

	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("TransferCargoToNearbyTransport: expected a blackboard")
		return bt.Running
	}

	l := bb.Logger().With("behavior", "TransferCargoToNearbyTransport")

	transporters := bb.mission.GetShipsByRole(MissionShipRoleHauler)
	var transport *Ship
	for _, t := range transporters {
		isAvailable := t.At(bb.extractionWaypoint) && !t.InTransit() && !t.IsCargoFull()
		if isAvailable {
			transport = t
			break
		}
		// l.Debug(fmt.Sprintf("transport %s not available", t.symbol), "in-transit", t.InTransit(), "fullCargo", t.IsCargoFull())
	}

	if transport == nil {
		return bt.Failure
	}

	l = l.With("transport.ship", transport.symbol, "transport.cargo.cap", transport.state.Cargo.Capacity, "transport.cargo.free", transport.AvailableCargoUnits())

	good := bb.ship.state.Cargo.Inventory[0]

	_, err := transport.ReceiveTransfer(bb.ship, good.Symbol, -1)
	if err != nil {
		l.Error(err.Error())
		return bt.Running
	}

	return bt.Running
}

type JettisonNonSellableCargo struct{}

func (a JettisonNonSellableCargo) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("JettisonNonSellableCargo: expected a blackboard")
		return bt.Running
	}

	for _, inv := range bb.ship.state.Cargo.Inventory {
		markets, err := bb.repo.FindMarketsForGoods([]string{string(inv.Symbol)})
		if err != nil {
			bb.Logger().Error(errors.Wrap(err, "JettisonNonSellableCargo failed:").Error())
			return bt.Running
		}

		if len(markets) == 0 || inv.Symbol == "QUARTZ_SAND" {
			bb.ship.Jettison(string(inv.Symbol))
		}
	}

	return bt.Success
}

type ExtractAction struct{}

func (a ExtractAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ExtractAction: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.Extract(api.Survey{}); err != nil {
		bb.Logger().Error(errors.Wrap(err, "Failed to extract").Error())
		return bt.Running
	}

	return bt.Running
}

type SellCargoAction struct{}

func (a SellCargoAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("SellCargoAction: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.SellCargo(); err != nil {
		bb.Logger().Error(errors.Wrap(err, "Failed to sell cargo").Error())
		return bt.Running
	}
	return bt.Success
}

type BuyAction struct{}

func (a BuyAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("BuyAction: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.Buy(bb.purchaseTargetGood, bb.purchaseMaxUnits, bb.destination); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to buy goods").Error())
		return bt.Failure
	}
	return bt.Success
}

type DeliverContractAction struct{}

func (a DeliverContractAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("DeliverContractAction: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("DeliverContractAction: blackboard: contract was nil")
		return bt.Running
	}
	if bb.ship == nil {
		slog.Error("DeliverContractAction: blackboard: ship was nil")
		return bt.Running
	}

	if !bb.ship.HasGood(bb.contract.Terms.Deliver[0].TradeSymbol) {
		return bt.Success
	}

	contract, err := bb.ship.Deliver(bb.contract.ID, bb.contract.Terms.Deliver[0].TradeSymbol)
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to deliver contract goods").Error())
		bb.ship.log.Debug("DeliverContractAction: fail")
		return bt.Failure
	}

	bb.contract.Contract = contract

	return bt.Success
}

type TodoBehavior struct {
	name string
}

func NewTodoBehavior(name string) TodoBehavior {
	return TodoBehavior{name}
}

func (a TodoBehavior) Tick(data any) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("TodoBehavior: expected a blackboard")
		return bt.Running
	}
	bb.Logger().Warn("TODO:"+a.name, "behavior", "TodoBehavior")
	return bt.Running
}

type ConditionHasNonContractGoods struct{}

func (a ConditionHasNonContractGoods) Tick(data any) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionHasNonContractGoods: expected a blackboard")
		return bt.Running
	}
	if bb.ship == nil {
		bb.Logger().Error("ConditionHasNonContractGoods: blackboard: ship was nil")
		return bt.Running
	}
	if bb.contract == nil {
		bb.Logger().Error("ConditionHasNonContractGoods: blackboard: contract was nil")
		return bt.Running
	}

	contractGoods := []string{}
	for _, d := range bb.contract.Terms.Deliver {
		contractGoods = append(contractGoods, d.TradeSymbol)
	}

	for _, inv := range bb.ship.state.Cargo.Inventory {
		if !slices.Contains(contractGoods, string(inv.Symbol)) {
			return bt.Success
		}
	}

	return bt.Failure
}

type ConditionHasCargo struct{}

func (a ConditionHasCargo) Tick(data any) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionHasCargo: expected a blackboard")
		return bt.Running
	}
	if bb.ship == nil {
		bb.Logger().Error("ConditionHasCargo: blackboard: ship was nil")
		return bt.Running
	}

	if bb.ship.IsCargoEmpty() {
		return bt.Failure
	}

	return bt.Success
}

type WaitForCargoFull struct{}

func (a WaitForCargoFull) Tick(data any) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("WaitForCargoFull: expected a blackboard")
		return bt.Running
	}
	if bb.ship == nil {
		slog.Error("WaitForCargoFull: blackboard: ship was nil")
		return bt.Running
	}

	l := bb.ship.log.With("cargo.avail", bb.ship.AvailableCargoUnits())
	if bb.ship.IsCargoFull() {
		l.Info("WaitForCargoFull: full!")
		return bt.Failure
	}

	l.Info("WaitForCargoFull: waiting for cargo")

	return bt.Running
}

type ConditionCargoIsFull struct{}

func (a ConditionCargoIsFull) Tick(data any) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionCargoIsFull: expected a blackboard")
		return bt.Running
	}
	if bb.ship == nil {
		slog.Error("ConditionCargoIsFull: blackboard: ship was nil")
		return bt.Running
	}

	// l := bb.ship.log.With("cargo.avail", bb.ship.AvailableCargoUnits())
	if bb.ship.IsCargoFull() {
		// l.Debug("ConditionCargoIsFull: success!")
		return bt.Success
	}
	// l.Debug("ConditionCargoIsFull: fail!")
	return bt.Failure
}

type ConditionContractAccepted struct{}

func (a ConditionContractAccepted) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionContractAccepted: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionContractAccepted: blackboard: contract was nil")
		return bt.Running
	}

	if bb.contract.Accepted {
		return bt.Success
	}

	return bt.Failure
}

type AcceptContractAction struct{}

func (a AcceptContractAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("AcceptContractAction: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("AcceptContractAction: blackboard: contract was nil")
		return bt.Running
	}
	if bb.ship == nil {
		slog.Error("AcceptContractAction: blackboard: ship was nil")
		return bt.Running
	}

	res, err := bb.ship.client.AcceptContract(context.TODO(), api.AcceptContractParams{ContractId: bb.contract.ID})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to accept contract").Error())
		slog.Debug("AcceptContractAction: fail")
		return bt.Running
	}

	bb.contract.Contract = &res.Data.Contract

	slog.Debug("AcceptContractAction: success")
	return bt.Success
}

type ConditionCanFulfillContract struct {
	contract *Contract
}

func NewConditionCanFulfillContract(contract *Contract) *ConditionCanFulfillContract {
	return &ConditionCanFulfillContract{contract}
}

func (a *ConditionCanFulfillContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	if a.contract.Fulfilled {
		return bt.Success
	}
	return bt.Failure
}

type ConditionContractTermsMet struct{}

func (a ConditionContractTermsMet) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionContractTermsMet: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionContractTermsMet: blackboard: contract was nil")
		return bt.Running
	}

	for _, d := range bb.contract.Terms.Deliver {
		if d.UnitsRequired-d.UnitsFulfilled == 0 {
			continue
		}
		return bt.Failure
	}

	return bt.Success
}

type FulfillContractAction struct{}

func (a FulfillContractAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("PurchaseContractGoodsAction: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionContractTermsMet: blackboard: contract was nil")
		return bt.Running
	}

	res, err := bb.ship.client.FulfillContract(context.TODO(), api.FulfillContractParams{ContractId: bb.contract.ID})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to fulfill contract").Error())
		return bt.Running
	}

	bb.contract.Contract = &res.Data.Contract

	return bt.Success
}

type IsAtWaypointAction struct {
	ship     *Ship
	waypoint string
}

func NewIsAtWaypointAction(ship *Ship, waypoint string) *IsAtWaypointAction {
	return &IsAtWaypointAction{ship, waypoint}
}

func (a *IsAtWaypointAction) Tick(data bt.Blackboard) bt.BehaviorStatus {
	if string(a.ship.state.Nav.WaypointSymbol) != a.waypoint {
		slog.Debug("NewIsAtWaypointAction: fail", "waypoint", a.waypoint)
		return bt.Failure
	}
	slog.Debug("NewIsAtWaypointAction: success", "waypoint", a.waypoint)
	return bt.Success
}
