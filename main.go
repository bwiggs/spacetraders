package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bot"
	"github.com/bwiggs/spacetraders-go/client"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
)

var r *repo.Repo

func init() {
	logHandler := tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})
	// logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	slog.SetDefault(logger)

	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	dburl := viper.GetString("DB")
	slog.Info("connecting to data repo: " + dburl)
	var err error
	r, err = repo.NewRepo(dburl)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func main() {
	tasks.Start()
	exec()
	run()
}

type TokenProvider struct{}

func (tp TokenProvider) AgentToken(ctx context.Context, operationName string) (api.AgentToken, error) {
	return api.AgentToken{Token: viper.GetString("API_TOKEN")}, nil
}

func exec() {
	client, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	go initBackgroundTasks(client)

	bot.Start(client, r)
}

func shutdown() {
	slog.Info("cleaning up")
	tasks.Stop()
	r.Close()
}

func initBackgroundTasks(client *api.Client) {
	tasks.SetInterval(func() {
		tasks.LogAgentMetrics(client)
	}, 1*time.Minute)

	tasks.SetInterval(func() {
		tasks.ScanMarkets(client, r, "X1-HK42")
		tasks.ScanShipyards(client, r, "X1-HK42")
	}, 5*time.Minute)
}

func run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		slog.Info("\nReceived SIGINT (Ctrl+C).")
		shutdown()
		os.Exit(0)
	}()

	select {}
}

func mineAsteroid(client *api.Client) {
	ctx := context.Background()

	contracts, err := client.GetContracts(ctx, api.GetContractsParams{})
	if err != nil {
		log.Fatal(err)
	}

	ct := models.NewContractManager(client, contracts.Data[0])

	ships, err := client.GetMyShips(ctx, api.GetMyShipsParams{})
	if err != nil {
		log.Fatal(err)
	}

	for _, ship := range ships.Data {
		if ship.Registration.Role == "EXCAVATOR" || ship.Registration.Role == "COMMAND" {
			ct.AssignShip(&ship)
		}
	}

	tasks.SetInterval(ct.Update, 20*time.Second)
}
