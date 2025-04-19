package main

import (
	"image/color"
	"log"
	"math"

	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/viper"
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
	antialias = true
	minZoom   = 1.2
	maxZoom   = 25.0
)

var waypoints []models.Waypoint
var backgroundColor = color.RGBA{R: 0, G: 9, B: 22, A: 255}

func main() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	defaultFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 16)
	hudFont = loadFont("ui/assets/IBMPlexMono-Medium.ttf", 10)

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
