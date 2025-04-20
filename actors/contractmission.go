package actors

import (
	"fmt"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type ContractMission struct {
	*BaseMission
	contract *Contract
	bt       bt.BehaviorNode
}

func (m *ContractMission) String() string {
	deliver := m.contract.Terms.Deliver[0]
	return fmt.Sprintf("Contract: %s --> %s", deliver.TradeSymbol, deliver.DestinationSymbol)
}

func (m *ContractMission) Execute(data Blackboard) {
	data.contract = m.contract
	data.repo = m.repo
	m.bt.Tick(&data)
	if data.complete {
		// TODO: unassigned the ship so it can be used for something else
	}
}

func NewContractMission(client *api.Client, repo *repo.Repo, contract *api.Contract) *ContractMission {
	base := NewBaseMission(client, repo)
	base.name = "ContractMission"
	return &ContractMission{
		BaseMission: base,
		contract: &Contract{
			Contract: contract,
		},
		bt: bt.NewSelector(
			// success if the contract is fulfulled
			ConditionContractIsFulfilled{},

			// success if the contract can by filfilled
			bt.NewSequence(
				ConditionContractTermsMet{},
				FulfillContractAction{},
			),

			bt.NewSequence(
				// ConditionIsProfitableTradeRouteForContract{},
				// execute the contract
				bt.NewSequence(
					// sell off any non-contract goods
					bt.NewSelector(
						bt.Invert(ConditionHasNonContractGoods{}),
						// SellCargoSequence
						// bt.NewSequence(
						SellCargoAction{},
						// )
					),

					// buy contract good sourcing
					bt.NewSelector(
						ConditionCargoIsFull{},
						ConditionShipHasRemainingContractUnits{},
						bt.NewSequence(
							SetPurchaseFromContract{},
							NavigationAction(),
							DockAction{},
							BuyAction{},
						),
					),

					// contract delivery

					SetDeliveryDestFromContract{},

					bt.NewSelector(
						ConditionIsAtNavDest{},
						bt.NewSequence(
							RefuelAction{},
							OrbitAction{},
							NavAction{},
						),
					),

					DockAction{},

					bt.NewSelector(
						ConditionContractAccepted{},
						AcceptContractAction{},
					),

					DeliverContractAction{},
				),
			),
		),
	}
}
