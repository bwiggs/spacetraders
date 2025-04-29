package main

import (
	"context"
	"image/color"
	"log"
	"log/slog"
	"math"
	"os"
	"path"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/kernel"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"golang.org/x/image/font"

	"net/http"
	_ "net/http/pprof"
)

func init() {

	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	if viper.GetBool("PPROF") {
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("failed to start pprof server", "err", err)
			}
		}()
	}

	logHandler := tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})
	// logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	slog.SetDefault(logger)

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
	defaultSystemZoom                 = 0.95
	maxSystemZoom                     = 100.0
	transitionSystemToGalaxyZoomLevel = 0.012

	transitionGalaxyToSystemZoomLevel = 4.0
	minGalaxyZoom                     = 0.005

	showSystemLabelsAtZoom         = 0.3
	showSystemModeDetailsZoomLevel = 0.4
)

var credits int
var currSystem string
var waypoints []*models.Waypoint
var systems []models.System
var ships []*api.Ship
var contract *api.Contract
var agents []*api.Agent
var agentsBySystem map[string]*api.Agent
var systemCoords map[string][]float64
var constellationColors map[string]color.Color

func main() {

	kern, err := kernel.New()
	if err != nil {
		slog.Error("failed to create kernel", "err", err)
		return
	}

	kern.Start()

	currSystem = viper.GetString("SYSTEM")

	repo := kern.Repo()

	slog.Info("loading waypoints")

	waypoints, err = repo.GetNonOrbitalWaypoints(currSystem)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("loading systems")
	systems, err = repo.GetSystems()
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("loading agents")
	agents, err = repo.GetAgents()
	if err != nil {
		log.Fatal(err)
	}

	agentsBySystem = make(map[string]*api.Agent)
	for _, agent := range agents {
		agentsBySystem[agent.Headquarters[:7]] = agent
	}

	slog.Info("loading fleet")
	ships, err = repo.GetFleet()
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("loading contellation colors")
	systemCoords = make(map[string][]float64)
	constellationColors = make(map[string]color.Color)
	for _, system := range systems {
		systemCoords[system.Symbol] = []float64{float64(system.X), float64(system.Y)}
		if _, found := constellationColors[system.Constellation]; !found {
			constellationColors[system.Constellation] = color.NRGBA{R: uint8(rand.Intn(156) + 100), G: uint8(rand.Intn(156) + 100), B: uint8(rand.Intn(156) + 100), A: 255}
		}
	}

	slog.Info("calculating waypoint distances")
	// compute distance from center for each waypoint
	for i := 0; i < len(waypoints); i++ {
		dx := float64(waypoints[i].X) - 0.0
		dy := float64(waypoints[i].Y) - 0.0
		waypoints[i].Dist = math.Hypot(dx, dy)
	}

	slog.Info("starting tasks")
	tasks.Start()

	slog.Info("starting tasks: updatefleet")
	go tasks.SetInterval(func() {
		if err := tasks.UpdateFleet(kern.Client(), kern.Repo()); err != nil {
			slog.Error("ui: fleet update: failed to update fleet", "err", err)
		}

		ships, err = repo.GetFleet()
		if err != nil {
			log.Fatal(err)
		}

	}, 10*time.Second)

	slog.Info("starting tasks: my credits")
	go tasks.SetInterval(func() {
		res, err := kern.Client().GetMyAgent(context.TODO())
		if err != nil {
			slog.Error("ui: fleet update: failed to update fleet", "err", err)
		}
		credits = int(res.Data.Credits)
	}, 1*time.Minute)

	slog.Info("starting tasks: contracts")
	go tasks.SetInterval(func() {
		contract, err = tasks.GetLatestContract(kern.Client())
		if err != nil {
			slog.Error("ui: fleet update: failed to update fleet", "err", err)
		}
	}, 15*time.Second)

	slog.Info("loading fonts")
	defaultFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 14)
	hudFont = loadFont("ui/assets/IBMPlexMono-Regular.ttf", 14)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(2560, 1600)

	ebiten.SetTPS(60)
	ebiten.SetWindowTitle("spacetraders.io")
	slog.Info("spacetraders.io - UI", "system", currSystem, "agent", viper.GetString("AGENT"))
	if err := ebiten.RunGame(NewGame(kern)); err != nil {
		log.Fatal(err)
	}
}
