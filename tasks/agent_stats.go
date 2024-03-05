package tasks

import (
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func StartAgentStatsTask() {
	s.NewJob(
		gocron.DurationJob(
			1*time.Second,
		),
		gocron.NewTask(
			func() {
				slog.Info("update agent info")
			},
		),
	)
}
