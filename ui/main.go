package main

import (
	"image/color"
	"log"
	"math"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"golang.org/x/image/font"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

var defaultFont font.Face
var hudFont font.Face

const (
	antialias              = true
	defaultZoom            = 1.2
	systemToGalaxyThresh   = .15
	galaxyToSystemThresh   = 4
	defaultSystemZoon      = 1.1
	minZoom                = 0.014
	maxZoom                = 15.0
	showSystemLabelsAtZoom = 0.4
)

var currSystem string
var waypoints []models.Waypoint
var systems []models.System
var systemCoords map[string][]float64
var constellationColors map[string]color.Color
var backgroundColor = color.RGBA{R: 0, G: 9, B: 22, A: 255}

func main() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	defaultFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 12)
	hudFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 10)

	r, err := repo.GetRepo()
	if err != nil {
		log.Fatal(err)
	}

	currSystem = viper.GetString("SYSTEM")

	waypoints, err = r.GetWaypoints(currSystem)
	if err != nil {
		log.Fatal(err)
	}

	systems, err = r.GetSystems()
	if err != nil {
		log.Fatal(err)
	}

	systemCoords = make(map[string][]float64)
	constellationColors = make(map[string]color.Color)
	for _, system := range systems {
		systemCoords[system.Symbol] = []float64{float64(system.X), float64(system.Y)}
		if _, found := constellationColors[system.Constellation]; !found {
			constellationColors[system.Constellation] = color.NRGBA{R: uint8(rand.Intn(156) + 100), G: uint8(rand.Intn(156) + 100), B: uint8(rand.Intn(156) + 100), A: 255}
		}
	}

	// compute distance from center for each waypoint
	for i := 0; i < len(waypoints); i++ {
		dx := float64(waypoints[i].X) - 0.0
		dy := float64(waypoints[i].Y) - 0.0
		waypoints[i].Dist = math.Hypot(dx, dy)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(1680, 1020)

	// ebiten.SetTPS(30)
	ebiten.SetWindowTitle("spacetraders.io")
	if err := ebiten.RunGame(NewGame(r)); err != nil {
		log.Fatal(err)
	}
}
