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

func (m *TradeMission) Execute(data *Blackboard) {
	shipRole, _ := m.GetShipRole(data.ship.symbol)
	data.log = data.log.With("mission", "TradeMission", "role", shipRole)
	data.repo = m.repo
	data.mission = m

	m.GetShipBehavior(data.ship.symbol).Tick(data)
}

func NewTradeMission(client api.Invoker, repo *repo.Repo) *TradeMission {
	base := NewBaseMission(client, repo)
	base.name = "TradeMission"

	base.roleBehaviors[MissionShipRoleTrader] = bt.NewSelector(
		bt.NewSequence(
			ConditionHasActiveTrade{},
			bt.NewSelector(
				// offload any goods that aren't no the next trade
				bt.NewSequence(
					ConditionIsAtTradeSource{},
					bt.Invert(ConditionCargoIsFull{}),
					bt.Invert(ConditionHasTradeCargo{}),
					ActionDock{},
					ActionBuy{},
				),
				bt.NewSequence(
					ConditionIsAtTradeBuyer{},
					ConditionHasTradeCargo{},
					ActionDock{},
					ActionSellCargo{},

					// if anything prior fails, this wont run
					ActionAssignBestTrade{},
				),
				bt.NewSelector(
					bt.NewSequence(
						ConditionHasTradeCargo{},
						ActionSetDestinationTradeBuyer{},
						NavigationAction(),
					),
					bt.NewSequence(
						ActionSetDestinationTradeSource{},
						NavigationAction(),
					),
				),
			),
		),

		ActionAssignBestTrade{},
	)

	return &TradeMission{
		BaseMission: base,
	}
}
