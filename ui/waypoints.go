package main

import (
	"image/color"

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
