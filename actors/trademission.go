package actors

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type TradeMission struct {
	*BaseMission
}

func (m *TradeMission) AssignShip(role MissionShipRole, ship *Ship) {
	ship.SetMission(m)
	m.BaseMission.AssignShip(role, ship)
}

func (m *TradeMission) Execute(data Blackboard) {
	shipRole, _ := m.GetShipRole(data.ship.symbol)
	data.log = data.log.With("mission", "TradeMission", "role", shipRole)
	data.repo = m.repo
	data.mission = m

	m.GetShipBehavior(data.ship.symbol).Tick(&data)
}

func NewTradeMission(client *api.Client, repo *repo.Repo) *TradeMission {
	base := NewBaseMission(client, repo)
	base.name = "TradeMission"

	base.roleBehaviors[MissionShipRoleTrader] = bt.NewSelector(
		// SELL
		bt.NewSequence(
			JettisonNonSellableCargo{},
			ConditionHasCargo{},
			SetDestinationToBestTradeMarketSale{},
			bt.NewSequence(
				NavigationAction(),
				DockAction{},
				SellCargoAction{},
			),
		),

		// BUY
		bt.NewSequence(
			SetDestinationToBestTradeMarket{},
			bt.NewSequence(
				NavigationAction(),
				DockAction{},
				BuyAction{},
			),
		),
	)

	return &TradeMission{
		BaseMission: base,
	}
}
