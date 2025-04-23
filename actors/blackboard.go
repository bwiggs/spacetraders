package actors

import (
	"log/slog"

	"github.com/bwiggs/spacetraders-go/repo"
)

type Blackboard struct {
	repo               *repo.Repo
	contract           *Contract
	mission            Mission
	ship               *Ship
	purchaseTargetGood string
	destination        string
	purchaseMaxUnits   int

	extractionWaypoint string

	complete bool

	log *slog.Logger
}

func (bb *Blackboard) Logger() *slog.Logger {
	if bb.log != nil {
		return bb.log
	}
	return &slog.Logger{}
}
