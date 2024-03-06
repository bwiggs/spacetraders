package models

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/davecgh/go-spew/spew"
)

type ShipStatus int

// Define the states
const (
	ShipStatusUnknown ShipStatus = iota
	ShipStatusMining
	ShipStatusDelivering
	ShipStatusSelling
)

func (s ShipStatus) ToString() string {
	switch s {
	case 1:
		return "MINING"
	case 2:
		return "DELIVERING"
	case 3:
		return "SELLING"
	}
	return "UNKNOWN"
}

// Define the events
const (
	ShipEventCargoFilled = iota
	ShipEventCargoEmptied
	ShipEventCargoDelivered
	ShipEventArrived
)

type ShipAssignment struct {
	*Ship
	api.Contract

	Status ShipStatus
}

func NewShipAssignment(ship *api.Ship, contract api.Contract) *ShipAssignment {
	sm := ShipAssignment{
		Ship:     NewShip(ship),
		Contract: contract,
	}

	sm.calculateStatus()

	return &sm
}

func (m *ShipAssignment) calculateStatus() {
	if m.IsCargoFull() {
		if m.CountInventoryBySymbol(m.Contract.Terms.Deliver[0].TradeSymbol) > 0 {
			m.Status = ShipStatusDelivering
		} else {
			m.Status = ShipStatusSelling
		}
	} else if m.IsCargoEmpty() {
		m.Status = ShipStatusMining
	} else {
		isAtMiningLoc := m.Nav.WaypointSymbol == "X1-HK42-AC5C"
		if isAtMiningLoc {
			m.Status = ShipStatusMining
		} else {
			m.Status = ShipStatusSelling
		}
	}
}

func (m *ShipAssignment) Transition(event int) {
	lg := slog.With("ship", m.Symbol, "status", m.Status)
	lg.Info(fmt.Sprintf("%s transitioning", m.Symbol))
	switch m.Status {
	case ShipStatusUnknown:
		m.calculateStatus()
	case ShipStatusMining:
		if event == ShipEventCargoFilled {
			numContractTradeItems := m.CountInventoryBySymbol(m.Contract.Terms.Deliver[0].TradeSymbol)
			if numContractTradeItems > 0 {
				lg.Info(fmt.Sprintf("%s MINING -> DELIVERING", m.Symbol))
				m.Status = ShipStatusDelivering
			} else {
				lg.Info(fmt.Sprintf("%s MINING -> SELLING", m.Symbol))
				m.Status = ShipStatusSelling
			}
		}
	case ShipStatusDelivering:
		if event == ShipEventCargoDelivered {
			m.Status = ShipStatusSelling
		}
	case ShipStatusSelling:
		if event == ShipEventCargoEmptied {
			m.Status = ShipStatusMining
		}
	}
}

func (m *ShipAssignment) Update(client *api.Client) {

	ctx := context.TODO()

	client.CreateSurvey(ctx, api.CreateSurveyParams{ShipSymbol: "BWIGGS-4"})
	lg := slog.With("ship.symbol", m.Symbol, "mission.status", m.Status.ToString(), "ship.nav.status", m.Nav.Status, "ship.cargo.cap", m.Cargo.Capacity, "ship.cargo.units", m.Cargo.Units, "ship.nav.waypoint", m.Nav.WaypointSymbol)

	miningWaypoint := "X1-HK42-AC5C"
	currLoc := string(m.Ship.Nav.WaypointSymbol)
	isAtMiningLoc := currLoc == miningWaypoint
	isAtDeliveryLoc := string(m.Nav.WaypointSymbol) == m.Contract.Terms.Deliver[0].DestinationSymbol

	numContractTradeItems := m.CountInventoryBySymbol(m.Contract.Terms.Deliver[0].TradeSymbol)

	if m.InTransit() {
		lg.Debug(fmt.Sprintf("%s in transit", m.Ship.Symbol))
		return
	}

	if m.InCooldown() {
		lg.Debug(fmt.Sprintf("%s cooldown", m.Ship.Symbol))
		return
	}

	if !m.IsFuelFull() {

		if !m.IsDocked() {
			_, err := client.DockShip(ctx, api.DockShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
		}
		_, err := client.RefuelShip(ctx, api.NewOptRefuelShipReq(api.RefuelShipReq{Units: api.NewOptInt(m.Ship.Fuel.Capacity - m.Ship.Fuel.Current)}), api.RefuelShipParams{ShipSymbol: m.Symbol})
		if err != nil {
			lg.Error(err.Error())
			return
		}
	}

	switch m.Status {
	case ShipStatusMining:
		lg.Debug(fmt.Sprintf("%s MINING", m.Ship.Symbol))
		if m.IsCargoFull() {
			lg.Debug(fmt.Sprintf("%s transition to with cargo filled", m.Ship.Symbol))
			m.Transition(ShipEventCargoFilled)
			return
		}

		if !isAtMiningLoc {
			if m.IsDocked() {
				lg.Debug(fmt.Sprintf("%s moving to orbiting", m.Ship.Symbol))
				_, err := client.OrbitShip(ctx, api.OrbitShipParams{ShipSymbol: m.Symbol})
				if err != nil {
					lg.Error(err.Error())
					return
				}
			}
			navReq := api.NewOptNavigateShipReq(api.NavigateShipReq{WaypointSymbol: miningWaypoint})
			_, err := client.NavigateShip(ctx, navReq, api.NavigateShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
			lg.Info(fmt.Sprintf("%s navigating to mining location", m.Ship.Symbol))

			return
		}

		if m.IsDocked() {
			lg.Debug(fmt.Sprintf("%s moving to orbit", m.Ship.Symbol))
			_, err := client.OrbitShip(ctx, api.OrbitShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
			// getting some 400s after this, hoping 3 second delay fixes it
			time.Sleep(3 * time.Second)
		}

		lg.Info(fmt.Sprintf("%s extracting resources", m.Ship.Symbol))
		dat, err := client.ExtractResources(ctx, api.OptExtractResourcesReq{}, api.ExtractResourcesParams{ShipSymbol: m.Ship.Symbol})
		// _, err := client.ExtractResourcesWithSurvey(ctx, api.NewOptSurvey(api.Survey{Signature: "X1-HK42-AC5C-6218BB"}), api.ExtractResourcesWithSurveyParams{ShipSymbol: m.Symbol})
		if err != nil {
			lg.Error(err.Error())
			return
		}
		lg.Info(fmt.Sprintf("%s extracted %d %s", m.Ship.Symbol, dat.Data.Extraction.Yield.Units, dat.Data.Extraction.Yield.Symbol))
		return
	case ShipStatusDelivering:
		if m.Contract.Fulfilled {
			m.Status = ShipStatusSelling
			return
		}
		lg.Debug(fmt.Sprintf("%s DELIVERING", m.Ship.Symbol))
		if !isAtDeliveryLoc {
			lg.Debug(fmt.Sprintf("%s need to transit to contract delivery location", m.Ship.Symbol))
			if m.IsDocked() {
				_, err := client.OrbitShip(ctx, api.OrbitShipParams{ShipSymbol: m.Symbol})
				if err != nil {
					lg.Error(err.Error())
					return
				}
			}
			navReq := api.NewOptNavigateShipReq(api.NavigateShipReq{WaypointSymbol: m.Contract.Terms.Deliver[0].DestinationSymbol})
			_, err := client.NavigateShip(ctx, navReq, api.NavigateShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
			lg.Info(fmt.Sprintf("%s transitting to contract delivery location", m.Ship.Symbol))
			return
		}
		lg.Debug(fmt.Sprintf("%s at contract delivery location", m.Ship.Symbol))

		if m.IsDocked() {
			lg.Debug(fmt.Sprintf("%s docking", m.Ship.Symbol))
			_, err := client.DockShip(ctx, api.DockShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
		}

		lg.Debug(fmt.Sprintf("%s delivering", m.Ship.Symbol))
		// deliver goods
		contractParams := api.DeliverContractParams{ContractId: m.Contract.ID}
		payload := api.OptDeliverContractReq{}
		payload.SetTo(api.DeliverContractReq{Units: numContractTradeItems, ShipSymbol: m.Symbol, TradeSymbol: m.Contract.Terms.Deliver[0].TradeSymbol})
		_, err := client.DeliverContract(ctx, payload, contractParams)
		if err != nil {
			lg.Error(err.Error())
			return
		}
		lg.Info(fmt.Sprintf("%s cargo delivered", m.Ship.Symbol))
		m.Transition(ShipEventCargoDelivered)
		return
	case ShipStatusSelling:
		lg.Debug(fmt.Sprintf("%s SELLING", m.Ship.Symbol))
		if m.IsCargoEmpty() {
			lg.Debug(fmt.Sprintf("%s cargo empty, transitioning", m.Ship.Symbol))
			m.Transition(ShipEventCargoEmptied)
			return
		}

		lg.Debug(fmt.Sprintf("%s has cargo, checking markets", m.Ship.Symbol))

		resp, err := client.GetMarket(ctx, api.GetMarketParams{SystemSymbol: string(m.Nav.SystemSymbol), WaypointSymbol: string(m.Nav.WaypointSymbol)})
		if err != nil {
			lg.Error(err.Error())
			return
		}
		market := Market{Market: resp.Data}
		sellables := market.GetShipSellableGoods(m.Ship)
		if len(sellables) > 0 {

			if !m.IsDocked() {
				client.DockShip(ctx, api.DockShipParams{ShipSymbol: m.Symbol})
			}

			for i := range sellables {
				lg.Info(fmt.Sprintf("%s selling %d %s", m.Symbol, sellables[i].Units, sellables[i].Symbol))
				_, err := client.SellCargo(ctx, api.NewOptSellCargoReq(sellables[i]), api.SellCargoParams{ShipSymbol: m.Symbol})
				if err != nil {
					lg.Error(err.Error())
				}
			}
		}

		if m.IsCargoEmpty() {
			lg.Warn(fmt.Sprintf("%s cargo emptied transitioning", m.Ship.Symbol))
			m.Transition(ShipEventCargoEmptied)
			return
		}

		// TRANSITION TO A MARKET

		dest := "X1-HK42-H48"
		if m.Nav.WaypointSymbol == "X1-HK42-H48" {
			dest = "X1-HK42-H50"
		}
		slog.Debug(fmt.Sprintf("%s need to transit to %s", m.Ship.Symbol, dest))
		if m.IsDocked() {
			slog.Debug(fmt.Sprintf("%s undocking", m.Ship.Symbol))
			_, err := client.OrbitShip(ctx, api.OrbitShipParams{ShipSymbol: m.Symbol})
			if err != nil {
				lg.Error(err.Error())
				return
			}
		}
		slog.Debug(fmt.Sprintf("%s navigating", m.Ship.Symbol))
		navReq := api.NewOptNavigateShipReq(api.NavigateShipReq{WaypointSymbol: dest})
		_, err = client.NavigateShip(ctx, navReq, api.NavigateShipParams{ShipSymbol: m.Symbol})
		if err != nil {
			spew.Dump(err)
			lg.Error(err.Error())
			return
		}
		lg.Info(fmt.Sprintf("%s transitting to market %s", m.Ship.Symbol, dest))
		return
	default:
		lg.Debug(fmt.Sprintf("%s UNKNOWN", m.Ship.Symbol))
	}
}
