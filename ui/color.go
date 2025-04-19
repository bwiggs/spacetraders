package main

import (
	"image/color"
)

type GameColors struct {
	Background         color.Color
	Primary            color.Color
	Secondary          color.Color
	OrbitBaseColor     color.Color
	WaypointOrbit      color.Color
	DistanceRingsBase  color.Color
	DistanceRings      color.Color
	WaypointLabelColor color.Color
}

func fadeColorWithZoom(currZoom, zoomStart, zoomEnd, fromAlpha, toAlpha float64, c color.Color) color.Color {
	t := Clamp((currZoom-zoomStart)/(zoomEnd-zoomStart), 0.0, 1.0)
	a := fromAlpha + (toAlpha-fromAlpha)*t

	oc := colortoNRGBA(c)
	return color.NRGBA{oc.R, oc.G, oc.B, uint8(255 * a)}
}

func colortoNRGBA(c color.Color) color.NRGBA {
	r16, g16, b16, a16 := c.RGBA()

	// Convert from 16-bit (0–65535) to 8-bit (0–255)
	if a16 == 0 {
		return color.NRGBA{0, 0, 0, 0}
	}

	r := uint8((r16 * 0xFF) / a16)
	g := uint8((g16 * 0xFF) / a16)
	b := uint8((b16 * 0xFF) / a16)
	a := uint8(a16 >> 8)

	return color.NRGBA{R: r, G: g, B: b, A: a}
}
