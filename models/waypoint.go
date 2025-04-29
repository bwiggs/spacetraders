package models

import "github.com/bwiggs/spacetraders-go/api"

type Waypoint struct {
	*api.Waypoint
	Symbol string `db:"symbol"`
	Type   string `db:"type"`
	X      int    `db:"x"`
	Y      int    `db:"y"`

	// Dist is the distance from the sun, used for ui rendering
	Dist float64
}

func (w *Waypoint) CanRefuel() bool {
	for _, t := range w.Traits {
		if t.GetSymbol() == "MARKETPLACE" {
			return true
		}
	}
	return false
}

func NewWaypoint(wp *api.Waypoint) *Waypoint {
	return &Waypoint{
		Waypoint: wp,
		Symbol:   string(wp.GetSymbol()),
		Type:     string(wp.GetType()),
		X:        wp.X,
		Y:        wp.Y,
	}
}
