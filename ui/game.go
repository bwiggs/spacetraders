package main

import (
	"fmt"
	"image/color"
	"log/slog"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/spf13/viper"
	"golang.org/x/image/colornames"
)

type Game struct {
	camera *Camera2D

	mode ViewMode

	repo *repo.Repo
	// scale        float64
	cameraOffset [2]float64
	dragging     bool
	lastMousePos [2]int

	colors GameColors

	settings Settings
}

type Settings struct {
	showDistanceRings  bool
	showOrbitRings     bool
	showWaypointLabels bool
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
		repo:         r,
		cameraOffset: [2]float64{0, 0},
		settings: Settings{
			showDistanceRings:  true,
			showOrbitRings:     true,
			showWaypointLabels: true,
		},
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
		worldBeforeX, worldBeforeY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh)

		// 2. Apply zoom
		zoomFactor := 1 + scrollY*0.02
		newZoom := g.camera.Zoom * zoomFactor
		g.camera.Zoom = Clamp(newZoom, float64(minZoom), float64(maxZoom))

		// 3. World position under cursor after zoom
		worldAfterX, worldAfterY := g.camera.ScreenToWorld(mouseScreenX, mouseScreenY, sw, sh)

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
				sw, sh,
			)

			currWorldX, currWorldY := g.camera.ScreenToWorld(
				float64(x),
				float64(y),
				sw, sh,
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

	if inpututil.IsKeyJustReleased(ebiten.KeyF) {
		tps := ebiten.TPS()
		if tps == 30 {
			ebiten.SetTPS(60)
		} else if tps == 60 {
			ebiten.SetTPS(120)
		} else {
			ebiten.SetTPS(30)
		}
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyR) {
		g.settings.showDistanceRings = !g.settings.showDistanceRings
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyO) {
		g.settings.showOrbitRings = !g.settings.showOrbitRings
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyL) {
		g.settings.showWaypointLabels = !g.settings.showWaypointLabels
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyC) {
		slog.Info("Centering Camera")
		g.camera.LookAt(0, 0)
	} else if inpututil.IsKeyJustReleased(ebiten.KeyG) {
		slog.Info("Switching to Galaxy Mode")
		g.camera.LookAt(0, 0)
		g.mode = GalaxyMode
		g.camera.Zoom = minZoom
	} else if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		slog.Info("Switching to System Mode")
		g.camera.LookAt(0, 0)
		g.mode = SystemMode
		g.camera.Zoom = defaultSystemZoom
	} else {
		// update the camera mode based on zoom level

		if g.mode == GalaxyMode && g.camera.Zoom > galaxyToSystemThresh {
			slog.Info("Switching to System View")
			sw, sh := ebiten.WindowSize()

			// Step 1: Find where the galaxy coords appeared on screen
			gx := systemCoords[currSystem][0]
			gy := systemCoords[currSystem][1]

			screenX, screenY := g.camera.WorldToScreen(gx, gy, sw, sh)

			// Step 2: Switch to system mode
			g.mode = SystemMode
			g.camera.Zoom = systemToGalaxyThresh
			g.camera.LookAt(0, 0)

			// Step 3: Figure out what world position (in system coords) lives at that screen pixel
			// Remember: in system mode, world center is (0, 0), so we want *that* to appear at (screenX, screenY)
			wx, wy := g.camera.ScreenToWorld(screenX, screenY, sw, sh)

			// Step 4: Offset camera so that system-local (0, 0) maps to that pixel
			g.camera.LookAt(-wx, -wy)
		} else if g.mode == SystemMode && g.camera.Zoom < systemToGalaxyThresh {
			sw, sh := ebiten.WindowSize()

			// get current screen position of the system
			screenX, screenY := g.camera.WorldToScreen(0, 0, sw, sh)

			// switch modes
			g.mode = GalaxyMode
			g.camera.Zoom = galaxyToSystemThresh

			// look at system in galaxy space
			gx := systemCoords[currSystem][0]
			gy := systemCoords[currSystem][1]
			g.camera.LookAt(gx, gy)

			// get the world position of earlier screen pixel
			wx, wy := g.camera.ScreenToWorld(screenX, screenY, sw, sh)

			// reorient to position galaxy at that screen position
			g.camera.LookAt(gx+gx-wx, gy+gy-wy)
		}
	}

	if g.mode == SystemMode {
		g.colors.DistanceRings = fadeColorWithZoom(g.camera.Zoom, 0.75, 1.1, 0.1, .6, g.colors.Secondary)
		g.colors.WaypointOrbit = fadeColorWithZoom(g.camera.Zoom, 0.9, 1.1, 0, .1, g.colors.Primary)
		g.colors.WaypointLabelColor = fadeColorWithZoom(g.camera.Zoom, 1.0, 1.1, 0, 1, colornames.Silver)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(g.colors.Background)

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if g.mode == SystemMode {
		g.DrawWaypoint(screen, models.Waypoint{X: 0, Y: 0, Type: "STAR", Symbol: viper.GetString("SYSTEM")})
		g.DrawSystemUI(screen)
	} else {
		g.DrawGalaxyUI(screen)
	}

	g.DrawContractStatus(screen, nil)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%s | TPS: %.2f | ZOOM: %.3f | %dx%d", viper.GetString("SYSTEM"), ebiten.ActualTPS(), g.camera.Zoom, sw, sh))
}

func (g *Game) DrawSystemUI(screen *ebiten.Image) {
	if g.settings.showDistanceRings {
		g.DrawDistanceRings(screen)
	}
	if g.settings.showOrbitRings {
		g.DrawWaypointOrbits(screen, waypoints)
	}
	g.DrawWaypoints(screen, waypoints)
	// g.DrawWaypointList(screen, waypoints)
	g.DrawShips(screen, ships)
	g.DrawShipList(screen, ships)
}

func (g *Game) DrawGalaxyUI(screen *ebiten.Image) {
	g.DrawSystems(screen, systems)
}
