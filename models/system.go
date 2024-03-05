package models

import "github.com/bwiggs/spacetraders-go/api"

type Waypoint struct {
	api.Waypoint
}

type System struct {
	api.System

	Markets []Waypoint
}

func (s *System) GetClosestWaypointForTrade(origin api.Waypoint, item api.TradeSymbol) {

}
