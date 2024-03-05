package models

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
)

type ContractManager struct {
	client *api.Client

	Contract api.Contract
	Ships    []*ShipAssignment
}

func NewContractManager(client *api.Client, contract api.Contract) *ContractManager {
	return &ContractManager{
		client: client,

		Contract: contract,
		Ships:    []*ShipAssignment{},
	}
}

func (ct *ContractManager) AssignShip(ship *api.Ship) {
	slog.Info(fmt.Sprintf("Assigning ship %s (%s) to contract", ship.Symbol, ship.Registration.Role))
	ct.Ships = append(ct.Ships, NewShipAssignment(ship, ct.Contract))
}

func (ct *ContractManager) Update() {
	ctx := context.TODO()
	res, err := ct.client.GetMyShips(ctx, api.GetMyShipsParams{})
	if err != nil {
		slog.Error(err.Error())
		return
	}

	for _, ship := range ct.Ships {
		for _, s := range res.Data {
			if s.Symbol == ship.Symbol {
				ship.Ship.Ship = &s
			}
		}
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		ship.Update(ct.client)
	}
}
