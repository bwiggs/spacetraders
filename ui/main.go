package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strings"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/spf13/viper"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var defaultFont font.Face

const (
	antialias = true
)

var waypoints []models.Waypoint
var backgroundColor = color.RGBA{R: 0, G: 9, B: 22, A: 255}

type Game struct {
	camera    *Camera2D
	worldSize float64

	repo *repo.Repo
	// scale        float64
	cameraOffset [2]float64
	dragging     bool
	lastMousePos [2]int
}

func NewGame(r *repo.Repo) *Game {
	g := &Game{
		camera:    NewCamera2D(),
		worldSize: 2000.0,
		repo:      r,
		// scale:        1.0, // default zoom level
		cameraOffset: [2]float64{0, 0},
	}

	return g
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
		zoomFactor := 1 + scrollY*0.1
		newZoom := g.camera.Zoom * zoomFactor
		g.camera.Zoom = Clamp(newZoom, 1.0, 15.0)

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
		g.camera.Zoom = 1.0
	}

	return nil
}

func getLabelColorForZoom(zoom float64) color.Color {
	fadeStart := 3.0
	fadeEnd := 2.0
	t := (zoom - fadeStart) / (fadeEnd - fadeStart)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	return color.NRGBA{255, 255, 255, uint8(255 * (1 - t))}
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(backgroundColor)

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	minX, maxX, minY, maxY := g.camera.GetWorldBounds(sw, sh, g.worldSize)

	g.DrawDistanceRings(screen)

	// draw background orbits
	for i := range waypoints {
		if waypoints[i].Type == "MOON" {
			continue
		}
		g.DrawWaypointOrbit(screen, waypoints[i])
	}

	labelColor := getLabelColorForZoom(g.camera.Zoom)
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

			text.Draw(screen, wp.Type, defaultFont, textX, textY, labelColor)
			text.Draw(screen, id, defaultFont, textX, textY+18, labelColor)
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%s | TPS: %.2f | ZOOM: %.2f | %dx%d", viper.GetString("SYSTEM"), ebiten.ActualTPS(), g.camera.Zoom, sw, sh))
}

func (g *Game) DrawDistanceRings(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	centerX, centerY := 0.0, 0.0

	alpha := 120
	if g.camera.Zoom < 2.0 {
		alpha = 30
	}

	orbitColor := rgbaToNrgba(colornames.Aqua, uint8(alpha))
	for i := 1.0; i < 9; i++ {
		radius := i * 100.0

		orbitX, orbitY := g.camera.WorldToScreen(centerX, centerY, sw, sh, g.worldSize)
		scale, _, _ := g.camera.GetTransform(sw, sh, g.worldSize)

		vector.StrokeCircle(
			screen,
			float32(orbitX),
			float32(orbitY),
			float32(radius*scale),
			2,
			orbitColor,
			antialias,
		)
	}

}

func rgbaToNrgba(c color.RGBA, alpha uint8) color.NRGBA {
	return color.NRGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: alpha,
	}
}

func (g *Game) DrawWaypointOrbit(screen *ebiten.Image, wp models.Waypoint) {

	if g.camera.Zoom < 2.0 {
		return
	}

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	centerX, centerY := 0.0, 0.0

	alpha := 6
	if g.camera.Zoom > 2.0 {
		alpha = 12
	}

	orbitColor := rgbaToNrgba(colornames.Aqua, uint8(alpha))

	dx := float64(wp.X) - centerX
	dy := float64(wp.Y) - centerY
	radius := math.Hypot(dx, dy)

	orbitX, orbitY := g.camera.WorldToScreen(centerX, centerY, sw, sh, g.worldSize)
	scale, _, _ := g.camera.GetTransform(sw, sh, g.worldSize)

	vector.StrokeCircle(
		screen,
		float32(orbitX),
		float32(orbitY),
		float32(radius*scale),
		2,
		orbitColor,
		antialias,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	defaultFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 16)

	r, err := repo.GetRepo()
	if err != nil {
		log.Fatal(err)
	}

	wps, err := r.GetWaypoints(viper.GetString("SYSTEM"))
	if err != nil {
		log.Fatal(err)
	}
	waypoints = wps

	// add the home star to the center
	waypoints = append(waypoints, models.Waypoint{X: 0, Y: 0, Type: "STAR", Symbol: viper.GetString("SYSTEM")})

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(1680, 1020)

	ebiten.SetTPS(30)
	ebiten.SetWindowTitle("spacetraders.io")
	if err := ebiten.RunGame(NewGame(r)); err != nil {
		log.Fatal(err)
	}
}

func loadFont(path string, size float64) font.Face {
	tt, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	ttf, err := opentype.Parse(tt)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	return face
}
