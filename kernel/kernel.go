package kernel

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bot"
	"github.com/bwiggs/spacetraders-go/client"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/bwiggs/spacetraders-go/tasks"
	"github.com/lmittmann/tint"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()

	logHandler := tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})
	// logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	slog.SetDefault(logger)
}

type Kernel struct {
	logger *slog.Logger
	client api.Invoker
	repo   *repo.Repo
	state  *State
}

func New() (*Kernel, error) {

	logger := slog.Default()

	client, err := client.GetClient()
	if err != nil {
		logger.Error("failed to create client", "err", err)
		return nil, err
	}

	dburl := viper.GetString("DB")
	logger.Info("connecting to data repo: " + dburl)
	r, err := repo.NewRepo(dburl)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	ships, err := r.GetFleet()
	if err != nil {
		slog.Error(errors.Wrap(err, "failed to get fleet").Error())
		return nil, err
	}

	shipsBySymbol := make(map[string]*api.Ship)
	for _, s := range ships {
		shipsBySymbol[s.Symbol] = s
	}

	return &Kernel{
		client: client,
		repo:   r,
		logger: logger,
		state:  NewState(),
	}, nil
}

func (k *Kernel) Logger() *slog.Logger {
	return k.logger
}

func (k *Kernel) Client() api.Invoker {
	return k.client
}

func (k *Kernel) Repo() *repo.Repo {
	return k.repo
}

func (k *Kernel) State() *State {
	return k.state
}

func (k *Kernel) Start() error {
	tasks.Start()
	go k.initBackgroundTasks(k.client)

	bot.Start(k.client, k.repo)
	return nil
}

func (k *Kernel) Stop() error {
	k.logger.Info("cleaning up")
	tasks.Stop()
	k.repo.Close()
	return nil
}

func (k *Kernel) initBackgroundTasks(client api.Invoker) error {
	tasks.SetInterval(func() {
		tasks.LogAgentMetrics(client)
	}, 1*time.Minute)

	// tasks.SetInterval(func() {
	// 	tasks.UpdateFleet(client, k.repo)
	// 	ships, err := k.repo.GetFleet()
	// 	if err != nil {
	// 		slog.Error(errors.Wrap(err, "failed to get fleet").Error())
	// 	}
	// 	k.state.ShipsBySymbol = make(map[string]*api.Ship)
	// 	k.state.Ships = []*api.Ship{}
	// 	for _, s := range ships {
	// 		k.state.ShipsBySymbol[s.Symbol] = &s
	// 		k.state.Ships = append(k.state.Ships, &s)
	// 	}

	// }, 30*time.Second)

	ScanMarkets := false
	if ScanMarkets {
		tasks.SetInterval(func() {
			slog.Info("task: scanning markets and shipyards")
			err := tasks.ScanMarkets(client, k.repo, viper.GetString("SYSTEM"))
			if err != nil {
				slog.Error(errors.Wrap(err, "failed to scan markets").Error())
			}
			err = tasks.ScanShipyards(client, k.repo, viper.GetString("SYSTEM"))
			if err != nil {
				slog.Error(errors.Wrap(err, "failed to scan shipyards").Error())
			}
		}, 15*time.Minute)
	}
	return nil
}

func (k *Kernel) Run() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		slog.Info("\nReceived SIGINT (Ctrl+C).")
		k.Stop()
		os.Exit(0)
	}()

	k.Start()

	select {}
}
