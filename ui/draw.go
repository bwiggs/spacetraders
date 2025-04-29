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

func (g *Game) DrawShips(screen *ebiten.Image, ships []*api.Ship) {
	// sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	// minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh)
	for i := range ships {

		if ships[i].Nav.Status != api.ShipNavStatusINTRANSIT {
			continue
		}

		g.DrawShip(screen, ships[i])
	}
}

func (g *Game) DrawShip(screen *ebiten.Image, ship *api.Ship) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	origin := Point2D{X: float64(ship.Nav.Route.Origin.X), Y: float64(ship.Nav.Route.Origin.Y)}
	destination := Point2D{X: float64(ship.Nav.Route.Destination.X), Y: float64(ship.Nav.Route.Destination.Y)}
	pos := Interpolate(origin, destination, ship.Nav.Route.DepartureTime, ship.Nav.Route.Arrival, time.Now())

	sx, sy := g.camera.WorldToScreen(pos.X, pos.Y, sw, sh)

	vector.DrawFilledRect(screen, float32(sx), float32(sy), 4, 4, g.colors.Secondary, g.settings.antialias)

	labelOffsetX := sx + 10.0
	labelOffset := float32(11)
	eta := time.Until(ship.Nav.Route.Arrival)
	text.Draw(screen, ship.Symbol, defaultFont, int(labelOffsetX), int(sy)+7, colornames.White)
	if eta > 0 {
		parts := strings.Split(ship.Nav.Route.Destination.Symbol, "-")
		name := parts[len(parts)-1]
		text.Draw(screen, name+" "+formatDuration(eta), defaultFont, int(labelOffsetX), int(sy)+20, colornames.White)
	}

	showBars := false

	if showBars {
		barWidth := float32(70.0)
		barHeight := float32(6.0)
		currOffsetY := float32(sy) + 20.0

		if ship.Fuel.Capacity > 0 {
			fuel := float32(ship.Fuel.Current/ship.Fuel.Capacity) * barWidth
			vector.StrokeRect(screen, float32(labelOffsetX), currOffsetY, barWidth, barHeight, 2, g.colors.Primary, g.settings.antialias)
			vector.DrawFilledRect(screen, float32(labelOffsetX), currOffsetY, fuel, barHeight, g.colors.Primary, g.settings.antialias)
			currOffsetY += labelOffset
		}

		if ship.Cargo.Capacity > 0 {
			cargo := float32(ship.Cargo.Units/ship.Cargo.Capacity) * barWidth
			vector.StrokeRect(screen, float32(labelOffsetX), currOffsetY, barWidth, barHeight, 2, g.colors.Primary, g.settings.antialias)
			vector.DrawFilledRect(screen, float32(labelOffsetX), currOffsetY, cargo, barHeight, g.colors.Primary, g.settings.antialias)
		}
	}
}

func (g *Game) DrawSystems(screen *ebiten.Image, systems []models.System) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh)
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
	sx, sy := g.camera.WorldToScreen(float64(system.X), float64(system.Y), sw, sh)

	size := float32(1)
	if g.camera.Zoom > .1 {
		size = 2
	} else if g.camera.Zoom > .2 {
		size = 3
	}

	// c := constellationColors[system.Constellation]
	// if c == nil {
	// 	c = colornames.White
	// }
	c, found := SystemTypeColors[system.Type]
	if !found {
		c = colornames.Aqua
	}
	if system.Symbol == currSystem {
		c = colornames.Lime
		size *= 2
	}

	vector.DrawFilledRect(screen, float32(sx), float32(sy), size, size, c, g.settings.antialias)
	// vector.StrokeLine(screen, float32(sx), float32(sy), float32(sx+size), float32(sy+size), size, c, g.settings.antialias)

	labelColor := colornames.White
	// if currSystem == system.Symbol {
	// 	text.Draw(screen, fmt.Sprintf("%s %s (%s)", agent.Symbol, system.Symbol, system.Name), defaultFont, int(sx)+10.0, int(sy)+7, labelColor)
	// } else
	if agent, found := agentsBySystem[system.Symbol]; found {
		if system.Symbol == currSystem {
			labelColor = colornames.Lime
		}
		text.Draw(screen, fmt.Sprintf("%s %s", agent.Symbol, system.Symbol), defaultFont, int(sx)+10.0, int(sy)+7, labelColor)
	} else if g.camera.Zoom > showSystemModeDetailsZoomLevel {
		text.Draw(screen, fmt.Sprintf("%s (%s)", system.Symbol, system.Name), defaultFont, int(sx)+10.0, int(sy)+7, colornames.Gold)
	}
}

func (g *Game) DrawWaypoints(screen *ebiten.Image, waypoints []*models.Waypoint) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh)

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

func (g *Game) DrawWaypoint(screen *ebiten.Image, waypoint *models.Waypoint) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	c, ok := WaypointTypeColors[waypoint.Type]
	if !ok {
		c = colornames.White
	}

	r := float32(3)
	if waypoint.Type == "STAR" {
		r = float32(5)
	} else if waypoint.Type == "ASTEROID" {
		r = 1.5
	}

	// draw waypoint
	sx, sy := g.camera.WorldToScreen(float64(waypoint.X), float64(waypoint.Y), sw, sh)
	if g.camera.Zoom < showSystemModeDetailsZoomLevel {
		vector.DrawFilledRect(screen, float32(sx), float32(sy), r, r, c, g.settings.antialias)
	} else {
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), max(1, r*float32(g.camera.Zoom)), c, g.settings.antialias)
	}

	// render waypoint label
	if g.camera.Zoom >= showSystemModeDetailsZoomLevel {

		textX := int(sx) + 10 + int(float64(r)*g.camera.Zoom) // shift text a bit right of the circle
		textY := int(sy) - 1                                  // shift text a bit up

		parts := strings.Split(waypoint.Symbol, "-")
		subtext := parts[len(parts)-1]
		if waypoint.Waypoint != nil {
			for _, o := range waypoint.Orbitals {
				parts := strings.Split(o.Symbol, "-")
				subtext += fmt.Sprintf(" %s", parts[len(parts)-1])
			}
		}

		if g.settings.showWaypointLabels {
			text.Draw(screen, waypoint.Type, defaultFont, textX, textY, g.colors.WaypointLabelColor)
			text.Draw(screen, subtext, defaultFont, textX, textY+13, g.colors.WaypointLabelColor)
		}
	}
}

func (g *Game) DrawDistanceRings(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	centerX, centerY := 0.0, 0.0

	for i := 1.0; i < 9; i++ {
		radius := i * 100.0

		orbitX, orbitY := g.camera.WorldToScreen(centerX, centerY, sw, sh)
		scale, _, _ := g.camera.GetTransform(sw, sh)

		vector.StrokeCircle(
			screen,
			float32(orbitX),
			float32(orbitY),
			float32(radius*scale),
			1,
			g.colors.DistanceRings,
			g.settings.antialias,
		)
	}

}

func (g *Game) DrawWaypointOrbits(screen *ebiten.Image, waypoints []*models.Waypoint) {
	if g.camera.Zoom < showSystemModeDetailsZoomLevel {
		return
	}
	for i := range waypoints {
		if waypoints[i].Type == "MOON" {
			continue
		}
		g.DrawWaypointOrbit(screen, waypoints[i])
	}
}

func (g *Game) DrawWaypointOrbit(screen *ebiten.Image, wp *models.Waypoint) {

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	orbitX, orbitY := g.camera.WorldToScreen(0.0, 0.0, sw, sh)
	scale, _, _ := g.camera.GetTransform(sw, sh)

	vector.StrokeCircle(
		screen,
		float32(orbitX),
		float32(orbitY),
		float32(wp.Dist*scale),
		2,
		g.colors.WaypointOrbit,
		g.settings.antialias,
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

func (g *Game) DrawShipList(screen *ebiten.Image, ships []*api.Ship) {
	x := 10
	y := 30
	for i := range ships {

		etastr := ""
		var sym string
		if ships[i].Nav.Status == api.ShipNavStatusINTRANSIT {
			sym = "→"
			eta := time.Until(ships[i].Nav.Route.Arrival)
			if eta > 0 {
				etastr = " " + formatDuration(eta)
			}
		} else if ships[i].Nav.Status == api.ShipNavStatusDOCKED {
			sym = "▣"
		} else if ships[i].Nav.Status == api.ShipNavStatusINORBIT {
			sym = "○"
		} else {
			sym = "?"
		}

		label := fmt.Sprintf("%-10s %-10s %s %-12s%s", ships[i].Symbol, string(ships[i].Registration.Role), sym, ships[i].Nav.GetWaypointSymbol(), etastr)
		text.Draw(screen, label, hudFont, x, y, colornames.Aqua)
		y += 15
	}
}

func (g *Game) DrawContracts(screen *ebiten.Image, contracts []api.Contract) {
	_, sh := g.WindowSize()
	contract := contracts[len(contracts)-1]
	isFulfilled := contract.GetFulfilled()

	lineHeight := 18
	numLines := 2

	contractValue := contract.Terms.GetPayment().OnAccepted + contract.Terms.GetPayment().OnFulfilled

	etx := time.Until(contract.Terms.GetDeadline())
	var status string
	var color color.Color
	if etx < 0 {
		status = "EXPIRED"
		color = colornames.Red
	} else if isFulfilled {
		status = "FULFILLED"
		color = colornames.Palegreen
	} else {
		color = colornames.Aqua
		status = formatDuration(etx)
		numLines += len(contract.Terms.Deliver)
	}
	label := fmt.Sprintf("%s - %s - $%d - %s", contract.FactionSymbol, contract.Type, contractValue, status)

	text.Draw(screen, "CONTRACTS", hudFont, 10, sh-(lineHeight*numLines), colornames.Orange)
	numLines--
	text.Draw(screen, label, hudFont, 10, sh-(lineHeight*numLines), color)
	numLines--

	if !isFulfilled {
		for _, td := range contract.Terms.Deliver {
			l := fmt.Sprintf("    %s %s %d/%d", td.DestinationSymbol, td.TradeSymbol, td.UnitsFulfilled, td.UnitsRequired)
			text.Draw(screen, l, hudFont, 10, sh-(lineHeight*numLines), colornames.Aqua)
			numLines--
		}
	}
}

func (g *Game) DrawCredits(screen *ebiten.Image, credits int) {
	sw, _ := g.WindowSize()
	text.Draw(screen, fmt.Sprintf("%d", credits), hudFont, sw-100, 30, colornames.Aqua)
}
