package main

import (
	"image/color"
)

var SystemTypeColors = make(map[string]color.Color)

func init() {
	// color created using https://observablehq.com/@shan/oklab-color-wheel, lightness 0.87, 9 slices
	SystemTypeColors["BLACK_HOLE"] = color.NRGBA{190, 234, 0, 255}
	SystemTypeColors["BLUE_STAR"] = color.NRGBA{0, 225, 255, 255}
	SystemTypeColors["ORANGE_STAR"] = color.NRGBA{255, 187, 0, 255}
	SystemTypeColors["RED_STAR"] = color.NRGBA{255, 139, 49, 255}
	SystemTypeColors["YOUNG_STAR"] = color.NRGBA{0, 255, 255, 255}
	SystemTypeColors["WHITE_DWARF"] = color.NRGBA{255, 255, 255, 255}
	SystemTypeColors["UNSTABLE"] = color.NRGBA{255, 125, 203, 255}
	SystemTypeColors["NEUTRON_STAR"] = color.NRGBA{0, 255, 155, 255}
	SystemTypeColors["HYPERGIANT"] = color.NRGBA{188, 184, 255, 255}
}
