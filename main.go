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

	ctx := context.Background()
	c, err := api.NewClient(viper.GetString("BASE_URL"), TokenProvider{})
	if err != nil {
		log.Fatal(err)
	}

	contracts, err := c.GetContracts(ctx, api.GetContractsParams{})
	if err != nil {
		log.Fatal(err)
	}

	ct := models.NewContractManager(c, contracts.Data[0])

	ships, err := c.GetMyShips(ctx, api.GetMyShipsParams{})
	if err != nil {
		log.Fatal(err)
	}

	for _, ship := range ships.Data {
		if ship.Registration.Role == "EXCAVATOR" || ship.Registration.Role == "COMMAND" {
			// if ship.Symbol == "BWIGGS-3" {
			ct.AssignShip(&ship)
		}
	}

	tasks.SetInterval(ct.Update, 15*time.Second)
	tasks.SetInterval(func() {
		tasks.LogAgentMetrics(c)
	}, 1*time.Minute)
}

func shutdown() {
	slog.Info("cleaning up")
	tasks.Stop()
	r.Close()
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
