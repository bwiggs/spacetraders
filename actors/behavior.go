package actors

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	algos "github.com/bwiggs/spacetraders-go/algos/routing"
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-faster/errors"
	"github.com/spf13/viper"
)

func NavigationAction() *bt.Selector {
	return bt.NewSelector(
		bt.NewSequence(
			ConditionIsAtNavDest{},
		),
		bt.NewSelector(
			bt.NewSequence(
				ConditionWaypointHasFuel{},
				bt.Invert(ConditionShipFuelFull{}),
				bt.AlwaysFail(RefuelAction{}),
			),
			bt.NewSequence(
				ActionUpdateMarkets(),
				OrbitAction{},
				NavAction{},
			),
		),
	)
}

func ActionUpdateMarkets() *bt.Sequence {
	return bt.NewSequence(
		bt.AlwaysSucceed(
			bt.NewSequence(
				NewConditionAtWaypointWithTrait("MARKETPLACE"),
				ActionDock{},
				ActionScanMarket{},
			),
		),
		bt.AlwaysSucceed(
			bt.NewSequence(
				NewConditionAtWaypointWithTrait("SHIPYARD"),
				ActionDock{},
				ActionScanShipyard{},
			),
		),
	)
}

type ConditionInTransitToDest struct{}

func (a ConditionInTransitToDest) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionInTransitToDest: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("ConditionInTransitToDest: blackboard: ship was nil")
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

type ConditionShipFuelFull struct{}

func (a ConditionShipFuelFull) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionShipFuelFull: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("ConditionShipFuelFull: blackboard: ship was nil")
		return bt.Running
	}

	if bb.ship.IsFuelFull() {
		return bt.Success
	}

	return bt.Failure
}

type ConditionWaypointHasFuel struct{}

func (a ConditionWaypointHasFuel) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionWaypointHasFuel: expected a blackboard")
		return bt.Running
	}

	if bb.ship == nil {
		slog.Error("ConditionWaypointHasFuel: blackboard: ship was nil")
		return bt.Running
	}

	wp := bb.ship.CurrWaypoint()

	bb.log.Debug("ConditionWaypointHasFuel: checking waypoint for fuel", "waypoint", wp)

	hasFuel, err := bb.repo.WaypointHasGood(wp, "FUEL")
	if err != nil {
		slog.Error(errors.Wrap(err, "ConditionWaypointHasFuel: failed to check waypoint for fuel").Error())
		return bt.Running
	}

	if hasFuel {
		return bt.Success
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

type ConditionHasActiveContract struct{}

func (a ConditionHasActiveContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasActiveContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract != nil && bb.contract.GetAccepted() {
		slog.Debug("ConditionHasActiveContract: true", "contract", bb.contract.ID)
		return bt.Success
	}

	slog.Debug("ConditionHasActiveContract: false")

	return bt.Failure
}

type ConditionNilContract struct{}

func (a ConditionNilContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionNilContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		return bt.Success
	}

	return bt.Failure
}

type ConditionContractExpired struct{}

func (a ConditionContractExpired) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionContractExpired: expected a blackboard")
		return bt.Running
	}

	if bb.contract != nil && bb.contract.IsExpired() {
		return bt.Success
	}

	return bt.Failure
}

type ActionClearContract struct{}

func (a ActionClearContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ActionClearContract: expected a blackboard")
		return bt.Running
	}

	bb.contract = nil

	return bt.Failure
}

type ConditionContractClosed struct{}

func (a ConditionContractClosed) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionContractClosed: expected a blackboard")
		return bt.Running
	}

	bb.log.Debug("ConditionContractClosed: status", "fulfilled", bb.contract.GetFulfilled(), "expired", bb.contract.IsExpired())

	if bb.contract.GetFulfilled() || bb.contract.IsExpired() {
		return bt.Success
	}

	return bt.Failure
}

type ConditionContractFulfilled struct{}

func (a ConditionContractFulfilled) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionContractFulfilled: expected a blackboard")
		return bt.Running
	}

	if bb.contract.GetFulfilled() {
		return bt.Success
	}

	return bt.Failure
}

type ConditionContractInProgress struct{}

func (a ConditionContractInProgress) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ConditionContractInProgress: expected a blackboard")
		return bt.Running
	}

	if bb.contract != nil && !bb.contract.IsExpired() && !bb.contract.GetFulfilled() && bb.contract.GetAccepted() {
		return bt.Success
	}

	return bt.Failure
}

type ActionSetLatestContract struct{}

func (a ActionSetLatestContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ActionSetLatestContract: expected a blackboard")
		return bt.Running
	}

	contract, err := tasks.GetLatestContract(bb.ship.client)
	if err != nil {
		bb.Logger().Error(errors.Wrap(err, "ActionSetLatestContract: failed to get latest contract").Error())
		return bt.Running
	}

	if bb.contract == nil {
		bb.contract = NewContract(contract)
		bb.Logger().Debug("ActionSetLatestContract: new contract", "contract", bb.contract.ID)
	}

	return bt.Success
}

type ConditionHasPendingContract struct{}

func (a ConditionHasPendingContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasPendingContract: expected a blackboard")
		return bt.Running
	}

	if bb.contract != nil && !bb.contract.GetAccepted() {
		slog.Debug("ConditionHasPendingContract: true", "contract", bb.contract.ID)
		return bt.Success
	}

	slog.Debug("ConditionHasPendingContract: false")

	return bt.Failure
}

type NegotiateNewContract struct{}

func (a NegotiateNewContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("NegotiateNewContract: expected a blackboard")
		return bt.Running
	}

	res, err := bb.ship.client.NegotiateContract(context.TODO(), api.NegotiateContractParams{ShipSymbol: bb.ship.symbol})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "NegotiateNewContract: Failed to negotiate contract").Error())
		return bt.Running
	}

	bb.contract = NewContract(&res.Data.Contract)

	return bt.Success
}

type AcceptContract struct{}

func (a AcceptContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("NegotiateNewContract: expected a blackboard")
		return bt.Running
	}

	res, err := bb.ship.client.AcceptContract(context.TODO(), api.AcceptContractParams{ContractId: bb.contract.ID})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "NegotiateNewContract: failed to accept contract").Error())
		return bt.Running
	}

	bb.contract = NewContract(&res.Data.Contract)

	return bt.Success
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
		slog.Info("ConditionContractIsFulfilled: contract complete!")
		bb.complete = true
		return bt.Success
	}

	return bt.Failure
}

type ConditionAtMarketplace struct{}

func (a ConditionAtMarketplace) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionAtMarketplace: expected a blackboard")
		return bt.Running
	}

	loc := bb.ship.CurrWaypoint()
	isMarketplace, err := bb.repo.WaypointHasTrait(loc, "MARKETPLACE")
	if err != nil {
		return bt.Running
	}

	if isMarketplace {
		return bt.Success
	}

	return bt.Failure
}

type ConditionAtWaypointWithTrait struct {
	trait string
}

func (a ConditionAtWaypointWithTrait) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionAtWaypointWithTrait: expected a blackboard")
		return bt.Running
	}

	loc := bb.ship.CurrWaypoint()
	hasTrait, err := bb.repo.WaypointHasTrait(loc, a.trait)
	if err != nil {
		return bt.Running
	}

	if hasTrait {
		return bt.Success
	}

	return bt.Failure
}

func NewConditionAtWaypointWithTrait(trait string) *ConditionAtWaypointWithTrait {
	return &ConditionAtWaypointWithTrait{trait: trait}
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
		l.Error(errors.Wrap(err, "SetDestinationToBestTradeMarketSale: FindMarketTrades failed:").Error())
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
		if remaining == 0 {
			continue
		}

		slog.Info("SetPurchaseFromContract: setting good to " + d.TradeSymbol)
		bb.purchaseTargetGood = d.TradeSymbol
		bb.purchaseMaxUnits = remaining

		// SET THE WAYPOINT TO BUY THE GOOD FROM

		markets, err := bb.repo.FindBuyWaypointsForGood(d.TradeSymbol)
		if err != nil {
			slog.Error(errors.Wrap(err, "SetPurchaseFromContract: FindMarketsWithGoods failed:").Error())
			return bt.Running
		}

		if len(markets) == 0 {
			slog.Warn("no markets available: " + d.TradeSymbol)
			return bt.Running
		}

		for _, m := range markets {
			if m == bb.ship.CurrWaypoint() {
				slog.Info("SetPurchaseFromContract: already at dest", "dest", m)
				bb.destination = m
				return bt.Success
			}
		}

		// figure out best path to the waypoint
		origin := bb.ship.CurrWaypoint()
		slog.Debug("SetPurchaseFromContract: finding best path/dest for good", "origin", origin)
		waypoints, err := bb.repo.GetWaypoints(bb.ship.System())
		if err != nil {
			slog.Error(errors.Wrap(err, "SetPurchaseFromContract: GetWaypoints failed:").Error())
			return bt.Running
		}

		market := markets[0]
		cost, path := algos.FindPath(bb.ship.state, market, waypoints)
		bb.log.Debug("SetPurchaseFromContract: found path to market", "origin", origin, "dest", market, "path", path, "cost", cost)
		if len(markets) > 1 {
			for i := 1; i < len(markets); i++ {
				c, p := algos.FindPath(bb.ship.state, markets[1], waypoints)
				bb.log.Debug("SetPurchaseFromContract: found path to market", "origin", origin, "dest", markets[1], "path", p, "cost", c)
				if c < cost {
					cost = c
					path = p
					market = markets[i]
				}
			}
		}

		if len(path) == 0 {
			slog.Error("SetPurchaseFromContract: no paths found for good")
			return bt.Running
		}

		bb.log.Debug("SetPurchaseFromContract: setting destination", "origin", origin, "dest", market, "path", path, "cost", cost)

		bb.destination = path[1]

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

type ConditionIsAtContractDestination struct{}

func (a ConditionIsAtContractDestination) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionIsAtContractDestination: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionIsAtContractDestination: blackboard: contract was nil")
		return bt.Running
	}

	if len(bb.contract.Contract.Terms.Deliver) > 1 {
		slog.Error("ConditionIsAtContractDestination: contract has multiple destinations")
		return bt.Running
	}

	if bb.ship.CurrWaypoint() == bb.contract.Contract.Terms.Deliver[0].DestinationSymbol {
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

	if time.Until(bb.ship.state.Nav.Route.Arrival) > 0 {
		return bt.Running
	}

	if bb.ship.At(bb.destination) {
		return bt.Success
	}

	// figure out best path to the waypoint
	waypoints, err := bb.repo.GetWaypoints(viper.GetString("SYSTEM"))
	if err != nil {
		slog.Error(errors.Wrap(err, "NavAction: GetWaypoints failed:").Error())
		return bt.Failure
	}

	cost, path := algos.FindPath(bb.ship.state, bb.destination, waypoints)
	if len(path) == 0 {
		slog.Error("NavAction: no path found")
		return bt.Failure
	}

	bb.log.Debug("NavAction: found path to market", "origin", bb.ship.CurrWaypoint(), "dest", bb.destination, "path", path, "cost", cost)

	bb.destination = path[1]

	if err := bb.ship.Transit(bb.destination); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to transit ship").Error())
		return bt.Failure
	}

	return bt.Running
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

type ActionDock struct{}

func (a ActionDock) Tick(data bt.Blackboard) bt.BehaviorStatus {
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

type ActionScanMarket struct{}

func (a ActionScanMarket) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ActionScanMarket: expected a blackboard")
		return bt.Running
	}

	cwp := bb.ship.CurrWaypoint()
	bb.ship.log.Info("ActionScanMarket: "+cwp, "waypoint", cwp)
	res, err := bb.ship.client.GetMarket(context.TODO(), api.GetMarketParams{SystemSymbol: cwp[:7], WaypointSymbol: cwp})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "ActionScanMarket: Failed to scan market").Error())
		return bt.Running
	}

	if err := bb.repo.UpsertMarket(res.Data); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "ActionScanMarket: Failed to upsert market").Error())
		return bt.Running
	}

	return bt.Success
}

type ActionScanShipyard struct{}

func (a ActionScanShipyard) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ActionScanShipyard: expected a blackboard")
		return bt.Running
	}

	cwp := bb.ship.CurrWaypoint()
	bb.ship.log.Info("ActionScanShipyard: "+cwp, "waypoint", cwp)
	res, err := bb.ship.client.GetShipyard(context.TODO(), api.GetShipyardParams{SystemSymbol: cwp[:7], WaypointSymbol: cwp})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "ActionScanShipyard: Failed to scan market").Error())
		return bt.Running
	}

	if err := bb.repo.UpsertShipyard(res.Data); err != nil {
		bb.ship.log.Error(errors.Wrap(err, "ActionScanShipyard: Failed to upsert market").Error())
		return bt.Running
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

type ActionSellCargo struct{}

func (a ActionSellCargo) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		bb.Logger().Error("ActionSellCargo: expected a blackboard")
		return bt.Running
	}

	if err := bb.ship.SellCargo(); err != nil {
		bb.Logger().Error(errors.Wrap(err, "ActionSellCargo: Failed to sell cargo").Error())
		return bt.Running
	}
	return bt.Success
}

type ActionBuy struct{}

func (a ActionBuy) Tick(data bt.Blackboard) bt.BehaviorStatus {
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

type ActionDeliverContractGoods struct{}

func (a ActionDeliverContractGoods) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ActionDeliverContractGoods: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ActionDeliverContractGoods: blackboard: contract was nil")
		return bt.Running
	}
	if bb.ship == nil {
		slog.Error("ActionDeliverContractGoods: blackboard: ship was nil")
		return bt.Running
	}

	if !bb.ship.HasGood(bb.contract.Terms.Deliver[0].TradeSymbol) {
		return bt.Success
	}

	contract, err := bb.ship.DeliverContract(bb.contract.ID, bb.contract.Terms.Deliver[0].TradeSymbol)
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "Failed to deliver contract goods").Error())
		bb.ship.log.Debug("ActionDeliverContractGoods: fail")
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

	if bb.ship.IsCargoFull() {
		return bt.Success
	}
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

type ConditionHasContractGoods struct{}

func (a ConditionHasContractGoods) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasContractGoods: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("ConditionHasContractGoods: blackboard: contract was nil")
		return bt.Running
	}

	contractGoods := []string{}
	for _, d := range bb.contract.Terms.Deliver {
		contractGoods = append(contractGoods, d.TradeSymbol)
	}

	for _, inv := range bb.ship.state.Cargo.Inventory {
		if slices.Contains(contractGoods, string(inv.Symbol)) {
			return bt.Success
		}
	}

	return bt.Failure
}

type ActionFulfillContract struct{}

func (a ActionFulfillContract) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("FulfillContractAction: expected a blackboard")
		return bt.Running
	}

	if bb.contract == nil {
		slog.Error("FulfillContractAction: blackboard: contract was nil")
		return bt.Running
	}

	res, err := bb.ship.client.FulfillContract(context.TODO(), api.FulfillContractParams{ContractId: bb.contract.ID})
	if err != nil {
		bb.ship.log.Error(errors.Wrap(err, "FulfillContractAction: failed to fulfill contract").Error())
		return bt.Running
	}

	slog.Debug("FulfillContractAction: success", "contract", res.Data.Contract.ID)

	bb.contract = nil

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

type ActionSleepNode struct {
	dur time.Duration
}

func (a ActionSleepNode) Tick(data bt.Blackboard) bt.BehaviorStatus {
	time.Sleep(a.dur)
	return bt.Success
}

func ActionSleep(duration time.Duration) ActionSleepNode {
	return ActionSleepNode{dur: duration}
}

type ConditionIsAtTradeSource struct{}

func (a ConditionIsAtTradeSource) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionIsAtTradeSource: expected a blackboard")
		return bt.Running
	}

	if bb.ship.CurrWaypoint() == bb.tradeSource {
		return bt.Success
	}

	return bt.Failure
}

type ConditionIsAtTradeBuyer struct{}

func (a ConditionIsAtTradeBuyer) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionIsAtTradeBuyer: expected a blackboard")
		return bt.Running
	}

	if bb.ship.CurrWaypoint() == bb.tradeBuyer {
		return bt.Success
	}

	return bt.Failure
}

type ConditionHasActiveTrade struct{}

func (a ConditionHasActiveTrade) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasActiveTrade: expected a blackboard")
		return bt.Running
	}

	if bb.tradeBuyer != "" {
		return bt.Success
	}

	return bt.Failure
}

type ActionAssignBestTrade struct{}

func (a ActionAssignBestTrade) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasActiveTrade: expected a blackboard")
		return bt.Running
	}

	trades, err := bb.repo.FindMarketTrades()
	if err != nil {
		bb.Logger().Error(errors.Wrap(err, "ActionAssignBestTrade: FindMarketTrades failed:").Error())
		return bt.Running
	}

	if len(trades) == 0 {
		bb.Logger().Info(errors.Wrap(err, "ActionAssignBestTrade: no trades found").Error())
		return bt.Running
	}

	bb.tradeSource = trades[0].Origin
	bb.tradeBuyer = trades[0].Dest
	bb.tradeGood = trades[0].Good

	// TODO: why limit this?
	bb.purchaseMaxUnits = 20

	return bt.Success
}

type ActionSetDestinationTradeBuyer struct{}

func (a ActionSetDestinationTradeBuyer) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ActionSetDestinationTradeBuyer: expected a blackboard")
		return bt.Running
	}

	bb.destination = bb.tradeBuyer

	return bt.Success
}

type ActionSetDestinationTradeSource struct{}

func (a ActionSetDestinationTradeSource) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ActionSetDestinationTradeSource: expected a blackboard")
		return bt.Running
	}

	bb.destination = bb.tradeSource

	return bt.Success
}

type ConditionHasTradeCargo struct{}

func (a ConditionHasTradeCargo) Tick(data bt.Blackboard) bt.BehaviorStatus {
	bb, ok := data.(*Blackboard)
	if !ok {
		slog.Error("ConditionHasTradeCargo: expected a blackboard")
		return bt.Running
	}

	spew.Dump(bb.ship.state.Cargo.Inventory)
	bb.Logger().Debug("ConditionHasTradeCargo: checking for trade cargo", "good", bb.tradeGood)

	if bb.ship.HasGood(bb.tradeGood) {
		return bt.Success
	}

	return bt.Failure
}
