package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

// Point2D represents a 2D position
type Point2D struct {
	X, Y float64
}

func Interpolate(origin, destination Point2D, startTime, endTime, currentTime time.Time) Point2D {
	totalDuration := endTime.Sub(startTime).Seconds()
	elapsed := currentTime.Sub(startTime).Seconds()

	progress := elapsed / totalDuration
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}

	return Point2D{
		X: origin.X + (destination.X-origin.X)*progress,
		Y: origin.Y + (destination.Y-origin.Y)*progress,
	}
}

func (g *Game) DrawShips(screen *ebiten.Image, ships []api.Ship) {
	// sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	// minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh, g.systemSize)
	for i := range ships {

		if ships[i].Nav.Status != api.ShipNavStatusINTRANSIT {
			continue
		}

		g.DrawShip(screen, ships[i])
	}
}

func (g *Game) DrawShip(screen *ebiten.Image, ship api.Ship) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	origin := Point2D{X: float64(ship.Nav.Route.Origin.X), Y: float64(ship.Nav.Route.Origin.Y)}
	destination := Point2D{X: float64(ship.Nav.Route.Destination.X), Y: float64(ship.Nav.Route.Destination.Y)}
	pos := Interpolate(origin, destination, ship.Nav.Route.DepartureTime, ship.Nav.Route.Arrival, time.Now())

	sx, sy := g.camera.WorldToScreen(pos.X, pos.Y, sw, sh, g.systemSize)

	vector.DrawFilledRect(screen, float32(sx), float32(sy), 4, 4, g.colors.Secondary, antialias)

	labelOffsetX := sx + 10.0
	text.Draw(screen, ship.Symbol, defaultFont, int(labelOffsetX), int(sy)+7, colornames.White)

	showBars := false

	if showBars {
		barWidth := float32(70.0)
		barHeight := float32(6.0)
		fuel := float32(ship.Fuel.Current/ship.Fuel.Capacity) * barWidth
		vector.StrokeRect(screen, labelOffsetX, float32(sy)+12, barWidth, barHeight, 2, g.colors.Primary, antialias)
		vector.DrawFilledRect(screen, labelOffsetX, float32(sy)+12, fuel, barHeight, g.colors.Primary, antialias)

		cargo := float32(ship.Cargo.Units/ship.Cargo.Capacity) * barWidth
		vector.StrokeRect(screen, labelOffsetX, float32(sy)+22, barWidth, barHeight, 2, g.colors.Primary, antialias)
		vector.DrawFilledRect(screen, labelOffsetX, float32(sy)+22, cargo, barHeight, g.colors.Primary, antialias)
	}
}

func (g *Game) DrawSystems(screen *ebiten.Image, systems []models.System) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh, g.systemSize)
	for i := range systems {
		// cull if it's offscreen
		wx := float32(systems[i].X)
		wy := float32(systems[i].Y)

		if wx < minX || wx > maxX || wy < minY || wy > maxY {
			continue
		}

		g.DrawSystem(screen, systems[i])
	}
}

func (g *Game) DrawSystem(screen *ebiten.Image, system models.System) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	sx, sy := g.camera.WorldToScreen(float64(system.X), float64(system.Y), sw, sh, g.systemSize)

	size := float32(1)
	if g.camera.Zoom > .1 {
		size = 2
	} else if g.camera.Zoom > .2 {
		size = 3
	}

	c := constellationColors[system.Constellation]
	if c == nil {
		c = colornames.White
	}
	if system.Symbol == currSystem {
		c = colornames.Lime
	}

	vector.DrawFilledRect(screen, float32(sx), float32(sy), size, size, c, antialias)
	if g.camera.Zoom > showSystemLabelsAtZoom {
		text.Draw(screen, fmt.Sprintf("%s (%s)", system.Symbol, system.Name), defaultFont, int(sx)+10.0, int(sy)+7, colornames.White)
	}
}

func (g *Game) DrawWaypoints(screen *ebiten.Image, waypoints []models.Waypoint) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh, g.systemSize)

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

		g.DrawWaypoint(screen, wp)
	}
}

func (g *Game) DrawWaypoint(screen *ebiten.Image, waypoint models.Waypoint) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	c, ok := WaypointTypeColors[waypoint.Type]
	if !ok {
		c = colornames.White
	}

	r := float32(1)
	if waypoint.Type == "STAR" {
		r = float32(4)
	} else if waypoint.Type == "PLANET" {
		r = float32(2)
	}

	// draw waypoint
	sx, sy := g.camera.WorldToScreen(float64(waypoint.X), float64(waypoint.Y), sw, sh, g.systemSize)
	vector.DrawFilledCircle(screen, float32(sx), float32(sy), r*float32(g.camera.Zoom), c, antialias)

	// render waypoint label
	if g.camera.Zoom > 1.0 {

		textX := int(sx) + 10 + int(float64(r)*g.camera.Zoom) // shift text a bit right of the circle
		textY := int(sy) - 1                                  // shift text a bit up

		parts := strings.Split(waypoint.Symbol, "-")
		id := parts[len(parts)-1]

		if g.settings.showWaypointLabels {
			text.Draw(screen, waypoint.Type, defaultFont, textX, textY, g.colors.WaypointLabelColor)
			text.Draw(screen, id, defaultFont, textX, textY+12, g.colors.WaypointLabelColor)
		}
	}
}

func (g *Game) DrawDistanceRings(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	centerX, centerY := 0.0, 0.0

	for i := 1.0; i < 9; i++ {
		radius := i * 100.0

		orbitX, orbitY := g.camera.WorldToScreen(centerX, centerY, sw, sh, g.systemSize)
		scale, _, _ := g.camera.GetTransform(sw, sh, g.systemSize)

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
	if g.camera.Zoom < 1.0 {
		return
	}
	for i := range waypoints {
		if waypoints[i].Type == "MOON" {
			continue
		}
		g.DrawWaypointOrbit(screen, waypoints[i])
	}
}

func (g *Game) DrawWaypointOrbit(screen *ebiten.Image, wp models.Waypoint) {

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	orbitX, orbitY := g.camera.WorldToScreen(0.0, 0.0, sw, sh, g.systemSize)
	scale, _, _ := g.camera.GetTransform(sw, sh, g.systemSize)

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

func (g *Game) DrawShipList(screen *ebiten.Image, ships []api.Ship) {
	x := 10
	y := 30
	for i := range ships {
		text.Draw(screen, ships[i].Symbol+" "+string(ships[i].Registration.Role), hudFont, x, y, colornames.Aqua)
		y += 12
	}
}

func (g *Game) DrawContractStatus(screen *ebiten.Image, contracts []api.Contract) {
	// sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

}
