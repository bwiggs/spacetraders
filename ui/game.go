package main

import (
	"fmt"
	"image/color"

	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/spf13/viper"
	"golang.org/x/image/colornames"
)

type Game struct {
	camera    *Camera2D
	worldSize float64

	repo *repo.Repo
	// scale        float64
	cameraOffset [2]float64
	dragging     bool
	lastMousePos [2]int

	colors GameColors
}

func NewGame(r *repo.Repo) *Game {
	g := &Game{
		camera:    NewCamera2D(),
		worldSize: 2000.0,
		repo:      r,
		// scale:        1.0, // default zoom level
		cameraOffset: [2]float64{0, 0},
		colors: GameColors{
			Background:    color.RGBA{R: 0, G: 9, B: 22, A: 255},
			Primary:       colornames.Aqua,
			Secondary:     colornames.Orange,
			WaypointOrbit: colornames.Aqua,
			DistanceRings: colornames.Aqua,
		},
	}

	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {

	if !ebiten.IsFocused() {
		return nil // Do nothing when the window isn't focused
	}

	// Zoom with scroll
	if _, scrollY := ebiten.Wheel(); scrollY != 0 {
		// Get screen size and cursor position
		sw, sh := ebiten.WindowSize()
		mx, my := ebiten.CursorPosition()
		mouseScreenX := float64(mx)
		mouseScreenY := float64(my)

		// 1. World position under cursor before zoom
		worldBeforeX, worldBeforeY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh, g.worldSize)

		// 2. Apply zoom
		zoomFactor := 1 + scrollY*0.02
		newZoom := g.camera.Zoom * zoomFactor
		g.camera.Zoom = Clamp(newZoom, float64(minZoom), float64(maxZoom))

		// 3. World position under cursor after zoom
		worldAfterX, worldAfterY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh, g.worldSize)

		// 4. Adjust camera center so world stays locked to cursor
		g.camera.CenterX += worldBeforeX - worldAfterX
		g.camera.CenterY += worldBeforeY - worldAfterY
	}

	// Drag to pan
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		sw, sh := ebiten.WindowSize()

		if !g.dragging {
			g.lastMousePos = [2]int{x, y}
			g.dragging = true
		} else {
			prevWorldX, prevWorldY := g.camera.ScreenToWorld(
				float64(g.lastMousePos[0]),
				float64(g.lastMousePos[1]),
				sw, sh, g.worldSize,
			)

			currWorldX, currWorldY := g.camera.ScreenToWorld(
				float64(x),
				float64(y),
				sw, sh, g.worldSize,
			)

			dx := prevWorldX - currWorldX
			dy := prevWorldY - currWorldY

			g.camera.CenterX += dx
			g.camera.CenterY += dy

			g.lastMousePos = [2]int{x, y}
		}
	} else {
		g.dragging = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.camera.CenterX = 0
		g.camera.CenterY = 0
		g.camera.Zoom = minZoom
	}

	g.colors.DistanceRings = fadeColorWithZoom(g.camera.Zoom, 1.2, 5.0, 0.1, .6, g.colors.Secondary)
	g.colors.WaypointOrbit = fadeColorWithZoom(g.camera.Zoom, 1.5, 5.0, 0, .1, g.colors.Primary)
	g.colors.WaypointLabelColor = fadeColorWithZoom(g.camera.Zoom, 2, 5.0, 0, 1, colornames.White)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(backgroundColor)

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	g.DrawDistanceRings(screen)
	g.DrawWaypointOrbits(screen, waypoints)
	g.DrawWaypoints(screen, waypoints)
	g.DrawWaypointList(screen, waypoints)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%s | TPS: %.2f | ZOOM: %.2f | %dx%d", viper.GetString("SYSTEM"), ebiten.ActualTPS(), g.camera.Zoom, sw, sh))
}
