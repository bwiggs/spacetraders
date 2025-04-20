package main

import (
	"fmt"
	"image/color"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/spf13/viper"
	"golang.org/x/image/colornames"
)

type Game struct {
	camera     *Camera2D
	systemSize float64
	galaxySize float64

	mode ViewMode

	repo *repo.Repo
	// scale        float64
	cameraOffset [2]float64
	dragging     bool
	lastMousePos [2]int

	colors GameColors
}

type ViewMode int

const (
	SystemMode ViewMode = iota
	GalaxyMode
)

func NewGame(r *repo.Repo) *Game {
	g := &Game{
		camera:       NewCamera2D(),
		mode:         SystemMode,
		systemSize:   2000.0,
		galaxySize:   2000.0,
		repo:         r,
		cameraOffset: [2]float64{0, 0},
		colors: GameColors{
			Background:         color.RGBA{R: 0, G: 9, B: 22, A: 255},
			WaypointLabelColor: color.NRGBA{R: 0, G: 0, B: 0, A: 0},
			Primary:            colornames.Aqua,
			Secondary:          colornames.Orange,
			WaypointOrbit:      colornames.Aqua,
			DistanceRings:      colornames.Aqua,
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
		worldBeforeX, worldBeforeY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh, g.systemSize)

		// 2. Apply zoom
		zoomFactor := 1 + scrollY*0.02
		newZoom := g.camera.Zoom * zoomFactor
		g.camera.Zoom = Clamp(newZoom, float64(minZoom), float64(maxZoom))

		// 3. World position under cursor after zoom
		worldAfterX, worldAfterY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh, g.systemSize)

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
				sw, sh, g.systemSize,
			)

			currWorldX, currWorldY := g.camera.ScreenToWorld(
				float64(x),
				float64(y),
				sw, sh, g.systemSize,
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
		g.camera.LookAt(0, 0)
	} else if ebiten.IsKeyPressed(ebiten.KeyG) {
		g.camera.LookAt(0, 0)
		g.mode = GalaxyMode
		g.camera.Zoom = minZoom
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camera.LookAt(0, 0)
		g.mode = SystemMode
		g.camera.Zoom = defaultSystemZoon
	} else {
		currMode := g.mode
		if g.mode == GalaxyMode && g.camera.Zoom > galaxyToSystemThresh {
			g.mode = SystemMode
			g.camera.Zoom = systemToGalaxyThresh
		} else if g.mode == SystemMode && g.camera.Zoom < systemToGalaxyThresh {
			g.mode = GalaxyMode
			g.camera.Zoom = galaxyToSystemThresh
		}

		if g.mode != currMode {
			// Reset camera position when switching modes
			if g.mode == GalaxyMode {
				sysCoords := systemCoords[currSystem]
				g.camera.LookAt(sysCoords[0], sysCoords[1])
			} else {
				g.camera.LookAt(0, 0)
			}
		}
	}

	if g.mode == SystemMode {
		g.colors.DistanceRings = fadeColorWithZoom(g.camera.Zoom, 1.0, 2.0, 0.1, .6, g.colors.Secondary)
		g.colors.WaypointOrbit = fadeColorWithZoom(g.camera.Zoom, 1.0, 2.0, 0, .1, g.colors.Primary)
		g.colors.WaypointLabelColor = fadeColorWithZoom(g.camera.Zoom, 1, 1.1, 0, 1, colornames.White)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(backgroundColor)

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if g.mode == SystemMode {
		g.DrawWaypoint(screen, models.Waypoint{X: 0, Y: 0, Type: "STAR", Symbol: viper.GetString("SYSTEM")})
		g.DrawSystemUI(screen)
	} else {
		g.DrawGalaxyUI(screen)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%s | TPS: %.2f | ZOOM: %.3f | %dx%d", viper.GetString("SYSTEM"), ebiten.ActualTPS(), g.camera.Zoom, sw, sh))
}

func (g *Game) DrawSystemUI(screen *ebiten.Image) {
	g.DrawDistanceRings(screen)
	g.DrawWaypointOrbits(screen, waypoints)
	g.DrawWaypoints(screen, waypoints)
	g.DrawWaypointList(screen, waypoints)
}

func (g *Game) DrawGalaxyUI(screen *ebiten.Image) {
	g.DrawSystems(screen, systems)
}
