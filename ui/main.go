package main

import (
	"fmt"
	"log"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/spf13/viper"
	"golang.org/x/image/colornames"
)

const (
	screenWidth  = 1400
	screenHeight = 1400
	antialias    = true
)

var waypoints []models.Waypoint

type Game struct {
	repo         *repo.Repo
	scale        float64
	cameraOffset [2]float64
	dragging     bool
	lastMousePos [2]int
}

func NewGame(r *repo.Repo) *Game {
	g := &Game{
		repo:         r,
		scale:        1.0, // default zoom level
		cameraOffset: [2]float64{0, 0},
	}

	return g
}

func (g *Game) Update() error {
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		mouseX, mouseY := ebiten.CursorPosition()
		cx := float64(mouseX)
		cy := float64(mouseY)

		oldScale := g.scale
		newScale := g.scale * (1 + scrollY*0.1)
		if newScale < 1.0 {
			newScale = 1.0
		}
		if newScale > 10 {
			newScale = 10
		}

		// Zoom to cursor logic
		g.cameraOffset[0] -= (cx - g.cameraOffset[0]) * (newScale/oldScale - 1)
		g.cameraOffset[1] -= (cy - g.cameraOffset[1]) * (newScale/oldScale - 1)

		g.scale = newScale
	}

	// --- Panning logic ---
	mouseX, mouseY := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !g.dragging {
			g.dragging = true
			g.lastMousePos = [2]int{mouseX, mouseY}
		} else {
			dx := mouseX - g.lastMousePos[0]
			dy := mouseY - g.lastMousePos[1]
			g.cameraOffset[0] += float64(dx)
			g.cameraOffset[1] += float64(dy)
			g.lastMousePos = [2]int{mouseX, mouseY}
		}
	} else {
		g.dragging = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.scale = 1.0
		g.cameraOffset = [2]float64{0, 0}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Helper to transform world coords to screen coords
	transform := func(wx, wy float32) (float32, float32) {
		x := (float64(wx)*g.scale + g.cameraOffset[0])
		y := (float64(wy)*g.scale + g.cameraOffset[1])
		return float32(x), float32(y)
	}

	// Draw center point (center of the system)
	cx, cy := transform(screenWidth/2, screenHeight/2)
	vector.DrawFilledCircle(screen, cx, cy, float32(4.0*g.scale), colornames.Gold, antialias)

	for i := range waypoints {
		c, ok := WaypointTypeColors[waypoints[i].Type]
		if !ok {
			c = colornames.White
		}

		r := float32(2.0 * g.scale)
		if waypoints[i].Type == "ASTEROID" {
			r = float32(1.0 * g.scale)
		}

		// Convert world coords to normalized coords (0..1), then scale to screen
		x := float32(waypoints[i].X+1000) / 2000 * screenWidth
		y := float32(waypoints[i].Y+1000) / 2000 * screenHeight
		sx, sy := transform(x, y)

		vector.DrawFilledCircle(screen, sx, sy, r, c, antialias)

		// render waypoint label
		if g.scale > 4.0 {
			label := waypoints[i].Symbol // or .Name if available
			textX := int(sx) + 25        // shift text a bit right of the circle
			textY := int(sy) - 10        // shift text a bit up
			ebitenutil.DebugPrintAt(screen, label, textX, textY)
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f | Scale: %.2f", ebiten.ActualTPS(), g.scale))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()
	r, err := repo.GetRepo()
	if err != nil {
		log.Fatal(err)
	}

	wps, err := r.GetWaypoints(viper.GetString("SYSTEM"))
	if err != nil {
		log.Fatal(err)
	}
	waypoints = wps

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetTPS(30)
	ebiten.SetWindowTitle("spacetraders.io")
	if err := ebiten.RunGame(NewGame(r)); err != nil {
		log.Fatal(err)
	}
}
