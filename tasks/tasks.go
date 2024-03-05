package tasks

import (
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var s gocron.Scheduler

func SetInterval(fn func(), t time.Duration) {
	s.NewJob(gocron.DurationJob(t), gocron.NewTask(fn))
	fn()
}

func Start() (err error) {
	slog.Debug("starting", "system", "Scheduler")
	s, err = gocron.NewScheduler()
	if err != nil {
		return nil
	}

	slog.Debug("registering tasks", "system", "Scheduler")
	// StartAgentStatsTask()

	s.Start()
	slog.Debug("started", "system", "Scheduler")

	return nil
}

func Stop() error {
	slog.Debug("stopping", "system", "Scheduler")
	if s == nil {
		return nil
	}
	return s.Shutdown()
}
