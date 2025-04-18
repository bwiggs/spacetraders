package actors

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type ExtractionMission struct {
	*BaseMission
	extractionWaypoint string
}

func (m *ExtractionMission) AssignShip(role MissionShipRole, ship *Ship) {
	ship.SetMission(m)
	m.BaseMission.AssignShip(role, ship)
}

func (m *ExtractionMission) Execute(data Blackboard) {
	shipRole, _ := m.GetShipRole(data.ship.symbol)
	data.log = data.log.With("mission", "ExtractionMission", "role", shipRole)
	data.extractionWaypoint = m.extractionWaypoint
	data.repo = m.repo
	data.mission = m

	m.GetShipBehavior(data.ship.symbol).Tick(&data)
}

func NewExtractionMission(client *api.Client, repo *repo.Repo, extractionWaypoint string) *ExtractionMission {
	base := NewBaseMission(client, repo)
	base.name = "ExtractionMission"

	gotoExtractionPoint := bt.NewSequence(
		bt.Invert(ConditionIsAtExtractionWaypoint{}),
		SetDestinationToExtractionWaypoint{},
		NavigationAction(),
	)

	extract := bt.NewSequence(
		ConditionIsAtExtractionWaypoint{},
		JettisonNonSellableCargo{},
		bt.Invert(ConditionCargoIsFull{}),
		ExtractAction{},
	)

	sell := bt.NewSequence(
		JettisonNonSellableCargo{},
		ConditionHasCargo{},
		SetDestinationToBestMarketToSellCargo{},
		bt.NewSequence(
			NavigationAction(),
			DockAction{},
			SellCargoAction{},
		),
	)

	transfer := bt.NewSequence(
		JettisonNonSellableCargo{},
		ConditionHasCargo{},
		TransferCargoToNearbyTransport{},
	)

	excavateBehavior := bt.NewSelector(
		gotoExtractionPoint,
		transfer,
		extract,
	)

	transportBehavior := bt.NewSelector(
		bt.NewSequence(
			ConditionIsAtExtractionWaypoint{},
			bt.Invert(ConditionCargoIsFull{}),
		),
		sell,
		gotoExtractionPoint,
	)

	excavateAndtransportBehavior := bt.NewSelector(
		extract,
		sell,
		gotoExtractionPoint,
	)

	base.roleBehaviors[MissionShipRoleExcavatorTransporter] = excavateAndtransportBehavior
	base.roleBehaviors[MissionShipRoleTransporter] = transportBehavior
	base.roleBehaviors[MissionShipRoleExcavator] = excavateBehavior

	return &ExtractionMission{
		BaseMission:        base,
		extractionWaypoint: extractionWaypoint,
	}
}
