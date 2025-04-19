package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

func (g *Game) DrawWaypoints(screen *ebiten.Image, waypoints []models.Waypoint) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh, g.worldSize)

	for _, wp := range waypoints {

		if wp.Type == "MOON" {
			continue
		}

		// cull if it's offscreen
		wx := float32(wp.X)
		wy := float32(wp.Y)

		if wx < minX || wx > maxX || wy < minY || wy > maxY {
			continue
		}

		c, ok := WaypointTypeColors[wp.Type]
		if !ok {
			c = colornames.White
		}

		r := float32(1)
		if wp.Type == "STAR" {
			r = float32(4)
		} else if wp.Type == "PLANET" {
			r = float32(2)
		}

		// draw waypoint
		sx, sy := g.camera.WorldToScreen(float64(wp.X), float64(wp.Y), sw, sh, g.worldSize)
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), r*float32(g.camera.Zoom), c, antialias)

		// render waypoint label
		if g.camera.Zoom > 2.0 {

			textX := int(sx) + 10 + int(float64(r)*g.camera.Zoom) // shift text a bit right of the circle
			textY := int(sy) - 1                                  // shift text a bit up

			parts := strings.Split(wp.Symbol, "-")
			id := parts[len(parts)-1]

			text.Draw(screen, wp.Type, defaultFont, textX, textY, g.colors.WaypointLabelColor)
			text.Draw(screen, id, defaultFont, textX, textY+18, g.colors.WaypointLabelColor)
		}
	}
}

func (g *Game) DrawDistanceRings(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	centerX, centerY := 0.0, 0.0

	for i := 1.0; i < 9; i++ {
		radius := i * 100.0

		orbitX, orbitY := g.camera.WorldToScreen(centerX, centerY, sw, sh, g.worldSize)
		scale, _, _ := g.camera.GetTransform(sw, sh, g.worldSize)

		vector.StrokeCircle(
			screen,
			float32(orbitX),
			float32(orbitY),
			float32(radius*scale),
			1,
			g.colors.DistanceRings,
			antialias,
		)
	}

}

func (g *Game) DrawWaypointOrbits(screen *ebiten.Image, waypoints []models.Waypoint) {
	for i := range waypoints {
		if waypoints[i].Type == "MOON" {
			continue
		}
		g.DrawWaypointOrbit(screen, waypoints[i])
	}
}

func (g *Game) DrawWaypointOrbit(screen *ebiten.Image, wp models.Waypoint) {

	if g.camera.Zoom < 1.5 {
		return
	}

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	orbitX, orbitY := g.camera.WorldToScreen(0.0, 0.0, sw, sh, g.worldSize)
	scale, _, _ := g.camera.GetTransform(sw, sh, g.worldSize)

	vector.StrokeCircle(
		screen,
		float32(orbitX),
		float32(orbitY),
		float32(wp.Dist*scale),
		2,
		g.colors.WaypointOrbit,
		antialias,
	)
}

func (g *Game) DrawWaypointList(screen *ebiten.Image, wp []models.Waypoint) {
	x := 10
	y := 30
	for i := range waypoints {
		text.Draw(screen, fmt.Sprintf("%s - %s", waypoints[i].Symbol, waypoints[i].Type), hudFont, x, y, color.White)
		y += 10
	}
}
