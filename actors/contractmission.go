package actors

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/bt"
	"github.com/bwiggs/spacetraders-go/repo"
)

type ContractMission struct {
	*BaseMission
	bt bt.BehaviorNode
}

func (m *ContractMission) String() string {
	return "ContractMission"
}

func (m *ContractMission) Execute(data *Blackboard) {
	data.repo = m.repo
	m.bt.Tick(data)
	if data.complete {
		// TODO: unassigned the ship so it can be used for something else
	}
}

func NewContractMission(client *api.Client, repo *repo.Repo) *ContractMission {
	base := NewBaseMission(client, repo)
	base.name = "ContractMission"
	return &ContractMission{
		BaseMission: base,
		bt: bt.NewSelector(

			bt.NewSequence(
				ConditionNilContract{},
				ActionSetLatestContract{},
			),

			bt.NewSequence(
				ConditionContractClosed{},
				DockAction{},
				NegotiateNewContract{},
			),

			bt.NewSequence(
				ConditionHasPendingContract{},
				// IsCurrentContractProfitable{},
				AcceptContract{},
			),

			// success if the contract can by filfilled
			bt.NewSequence(
				ConditionContractInProgress{},
				bt.NewSelector(

					bt.NewSequence(
						ConditionContractTermsMet{},
						FulfillContractAction{},
					),

					bt.NewSequence(
						// sell off any non-contract goods
						bt.NewSelector(
							bt.Not(ConditionHasNonContractGoods{}),
							// SellCargoSequence
							// bt.NewSequence(
							DockAction{},
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
		),
	}
}
