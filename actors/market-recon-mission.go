package actors

import (
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type MarketReconMission struct {
	*BaseMission
	bt bt.BehaviorNode
}

func (m *MarketReconMission) String() string {
	return "MarketReconMission"
}

func (m *MarketReconMission) Execute(data *Blackboard) {
	data.repo = m.repo
	m.bt.Tick(data)
	if data.complete {
		// TODO: unassigned the ship so it can be used for something else
	}
}

func NewMarketReconMission(client api.Invoker, repo *repo.Repo) *MarketReconMission {
	base := NewBaseMission(client, repo)
	base.name = "MarketReconMission"
	return &MarketReconMission{
		BaseMission: base,
		bt: bt.NewSequence(
			ActionUpdateMarkets(),
			ActionSleep(1*time.Minute),
		),
	}
}

func (m *MarketReconMission) AssignShip(role MissionShipRole, ship *Ship) {
	ship.SetMission(m)
	m.BaseMission.AssignShip(MissionShipRoleSatellite, ship)
}
