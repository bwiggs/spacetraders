package main

import (
	"image/color"

	"github.com/bwiggs/spacetraders-go/models"
	"golang.org/x/image/colornames"
)

var WaypointTypeColors = make(map[string]color.Color)

func init() {
	WaypointTypeColors["STAR"] = colornames.Gold
	WaypointTypeColors["PLANET"] = colornames.Dodgerblue
	WaypointTypeColors["MOON"] = colornames.Dodgerblue
	WaypointTypeColors["GAS_GIANT"] = colornames.Purple
	WaypointTypeColors["ORBITAL_STATION"] = colornames.Deeppink
	WaypointTypeColors["FUEL_STATION"] = colornames.Yellowgreen
	WaypointTypeColors["ENGINEERED_ASTEROID"] = colornames.Darkorange
	WaypointTypeColors["ASTEROID"] = colornames.Gray
	WaypointTypeColors["ASTEROID_BASE"] = colornames.Red
	WaypointTypeColors["JUMP_GATE"] = colornames.Lime
}

type Waypoint struct {
	*models.Waypoint
	X        int
	Y        int
	Type     string
	Label    string
	Sublabel string
}

func NewWaypoint(w *models.Waypoint) *Waypoint {
	return &Waypoint{
		Waypoint: w,
		X:        int(w.X),
		Y:        int(w.Y),
		Label:    w.Symbol,
		Type:     w.Type,
		Sublabel: w.Type,
	}
}
