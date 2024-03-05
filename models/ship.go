package models

import "github.com/bwiggs/spacetraders-go/api"

type Ship struct {
	*api.Ship
}

func NewShip(ship *api.Ship) *Ship {
	return &Ship{
		Ship: ship,
		// Receiver: make(chan Command)
	}
}

func (s *Ship) IsCargoFull() bool {
	return s.Ship.Cargo.Units == s.Ship.Cargo.Capacity
}

func (s *Ship) IsCargoEmpty() bool {
	return s.Ship.Cargo.Units == 0
}

func (s *Ship) IsDocked() bool {
	return s.Nav.Status == api.ShipNavStatusDOCKED
}

func (s *Ship) InTransit() bool {
	return s.Nav.Status == api.ShipNavStatusINTRANSIT
}

func (s *Ship) InCooldown() bool {
	return s.Cooldown.RemainingSeconds > 0
}

func (s *Ship) IsFuelFull() bool {
	return s.Fuel.Current == s.Fuel.Capacity
}

func (s *Ship) CountInventoryBySymbol(tradeSymbol string) int {
	for _, inv := range s.Cargo.Inventory {
		if inv.Symbol == api.TradeSymbol(tradeSymbol) {
			return inv.Units
		}
	}
	return 0
}
