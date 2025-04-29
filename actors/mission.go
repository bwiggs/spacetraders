package actors

import (
	"log/slog"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type MissionShipRole int

const (
	MissionShipRoleExcavator MissionShipRole = iota
	MissionShipRoleTransporter
	MissionShipRoleSurveyor
	MissionShipRoleExcavatorTransporter
	MissionShipRoleTrader
	MissionShipRoleHauler
	MissionShipRoleSatellite
)

type Mission interface {
	Execute(*Blackboard)

	AssignShip(MissionShipRole, *Ship)
	GetShipRole(string) (MissionShipRole, bool)
	GetShipsByRole(MissionShipRole) []*Ship
	GetShipBehavior(string) bt.BehaviorNode

	String() string
}

type ShipRoleMap map[string]MissionShipRole
type ShipsByRole map[MissionShipRole][]*Ship
type RoleBehaviors map[MissionShipRole]bt.BehaviorNode

func NewBaseMission(client api.Invoker, repo *repo.Repo) *BaseMission {
	return &BaseMission{
		client:        client,
		repo:          repo,
		shipRole:      make(ShipRoleMap),
		shipsByRole:   make(ShipsByRole),
		roleBehaviors: make(RoleBehaviors),
	}
}

type BaseMission struct {
	name          string
	client        api.Invoker
	repo          *repo.Repo
	roleBehaviors RoleBehaviors
	shipRole      ShipRoleMap
	shipsByRole   ShipsByRole
}

func (m *BaseMission) String() string {
	return m.name
}

func (m *BaseMission) GetShipRole(shipSymbol string) (MissionShipRole, bool) {
	r, found := m.shipRole[shipSymbol]
	return r, found
}

func (m *BaseMission) GetShipsByRole(role MissionShipRole) []*Ship {
	if ships, found := m.shipsByRole[role]; found {
		return ships
	}
	// return an empty list
	return []*Ship{}
}

func (m *BaseMission) GetShipBehavior(shipSymbol string) bt.BehaviorNode {
	role, ok := m.GetShipRole(shipSymbol)
	if !ok {
		return NewTodoBehavior("no behavior for role")
	}

	behavior, ok := m.roleBehaviors[role]
	if !ok {
		return NewTodoBehavior("no behavior for role")
	}

	return behavior
}

func (m *BaseMission) AssignShip(role MissionShipRole, ship *Ship) {
	m.shipRole[ship.symbol] = role

	ships, found := m.shipsByRole[role]
	if !found {
		ships = []*Ship{}
	}
	ships = append(ships, ship)
	m.shipsByRole[role] = ships
}

func (m *BaseMission) Execute(data *Blackboard) {
	slog.Info("BaseMission: execute")
}

func NewIdleMission() *IdleMission {
	return &IdleMission{
		name:        "IdleMission",
		BaseMission: NewBaseMission(nil, nil),
	}
}

type IdleMission struct {
	*BaseMission
	name string
}

func (m *IdleMission) Execute(ship *Ship) {
	ship.log.Info("sleeping")
}
