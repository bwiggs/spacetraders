package kernel

import "github.com/bwiggs/spacetraders-go/api"

type State struct {
	ShipsBySymbol map[string]*api.Ship
	Ships         []*api.Ship
	Systems       map[string]*api.System
	Waypoints     map[string]*api.Waypoint
	Contracts     map[string]*api.Contract
}

func NewState() *State {
	return &State{
		ShipsBySymbol: make(map[string]*api.Ship),
		Ships:         make([]*api.Ship, 0),
		Systems:       make(map[string]*api.System),
		Waypoints:     make(map[string]*api.Waypoint),
		Contracts:     make(map[string]*api.Contract),
	}
}

func (s *State) UpdateShip(ship *api.Ship) {
	s.ShipsBySymbol[ship.Symbol] = ship
	for i := range s.Ships {
		if s.Ships[i].Symbol == ship.Symbol {
			s.Ships[i] = ship
			break
		}
	}
}
