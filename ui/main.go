package main

import (
	"image/color"
	"log"
	"log/slog"
	"math"
	"os"
	"path"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"golang.org/x/image/font"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	logHandler := tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})
	// logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	slog.SetDefault(logger)
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("failed to start pprof server", "err", err)
		}
	}()

	ud, err := os.UserConfigDir()
	if err != nil {
		slog.Error("failed to get user config dir", "err", err)
		return
	}

	up := path.Join(ud, "spacetraders")
	if err := os.Mkdir(up, 0755); err != nil {
		if !os.IsExist(err) {
			slog.Error("failed to create config dir: "+up, "err", err)
			return
		}
	}

	slog.Info("config dir: " + up)
}

var defaultFont font.Face
var hudFont font.Face

const (
	antialias                      = false
	defaultSystemZoom              = 0.95
	minSystemZoom                  = 0.005
	maxSystemZoom                  = 15.0
	minGalaxyZoom                  = 0.011
	maxGalaxyZoom                  = 4
	showSystemLabelsAtZoom         = 0.4
	showSystemModeDetailsZoomLevel = 0.5
)

var currSystem string
var waypoints []models.Waypoint
var systems []models.System
var ships []api.Ship
var systemCoords map[string][]float64
var constellationColors map[string]color.Color

func main() {

	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	defaultFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 12)
	hudFont = loadFont("ui/assets/IBMPlexMono-Regular.ttf", 12)

	r, err := repo.GetRepo()
	if err != nil {
		log.Fatal(err)
	}

	currSystem = viper.GetString("SYSTEM")

	ships, err = r.GetFleet()
	if err != nil {
		log.Fatal(err)
	}

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
	ebiten.SetWindowSize(2560, 1600)

	ebiten.SetTPS(60)
	ebiten.SetWindowTitle("spacetraders.io")
	slog.Info("spacetraders.io - UI", "system", currSystem, "agent", viper.GetString("AGENT"))
	if err := ebiten.RunGame(NewGame(r)); err != nil {
		log.Fatal(err)
	}
}
