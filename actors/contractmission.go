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

func NewContractMission(client api.Invoker, repo *repo.Repo) *ContractMission {
	base := NewBaseMission(client, repo)
	base.name = "ContractMission"
	return &ContractMission{
		BaseMission: base,
		bt: bt.NewSelector(
			// success if the contract can by filfilled

			bt.NewSequence(
				ConditionContractInProgress{},

				bt.NewSelector(
					bt.NewSequence(
						ConditionIsAtContractDestination{},
						bt.NewSelector(
							bt.NewSequence(
								ConditionContractTermsMet{},
								ActionFulfillContract{},
							),
							bt.NewSequence(
								ConditionHasContractGoods{},
								ActionDock{},
								ActionDeliverContractGoods{},
							),

							bt.NewSequence(
								ConditionHasNonContractGoods{},
								bt.AlwaysFail(ActionSellCargo{}),
							),
						),
					),

					bt.NewSequence(
						ConditionCargoIsFull{},
						bt.NewSequence(
							SetDeliveryDestFromContract{},
							NavigationAction(),
						),
					),

					bt.NewSequence(
						SetPurchaseFromContract{},
						NavigationAction(),
						ActionDock{},
						ActionBuy{},
					),
				),
			),

			bt.NewSequence(
				ConditionNilContract{},
				ActionSetLatestContract{},
			),

			bt.NewSequence(
				bt.AlwaysFail(ActionSellCargo{}),
			),

			bt.NewSequence(
				ConditionContractClosed{},
				ActionDock{},
				NegotiateNewContract{},
			),

			bt.NewSequence(
				ConditionHasPendingContract{},
				// IsCurrentContractProfitable{},
				AcceptContract{},
			),
		),
	}
}
