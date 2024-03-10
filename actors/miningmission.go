package actors

import (
	"fmt"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
	"github.com/pkg/errors"
)

type MiningMission struct {
	origin  string
	state   State
	repo    *repo.Repo
	surveys []api.Survey
}

func NewMiningMission(r *repo.Repo, origin string) *MiningMission {
	return &MiningMission{
		origin: origin,
		state:  IdleState,
		repo:   r,
	}
}

func (m *MiningMission) String() string {
	return fmt.Sprintf("Mining %s", m.origin)
}

func (m *MiningMission) HasValidSurveys() bool {
	return m.surveys != nil && len(m.surveys) > 0
}

func (m *MiningMission) Execute(ship *Ship) {
	ship.log.Debug("MiningMission Executing", "state", m.state)

	cooldown := time.Until(ship.state.Cooldown.Expiration.Value)
	if cooldown > 0 {
		ship.log.Debug(fmt.Sprintf("cooldown: %s second", cooldown))
		time.Sleep(cooldown)
	}

	arrival := time.Until(ship.state.Nav.Route.Arrival)
	if arrival > 0 {
		ship.log.Debug(fmt.Sprintf("transiting: %s second", arrival))
		time.Sleep(arrival)
	}

	if m.state == IdleState {
		m.ComputeState(ship)
	}

	switch m.state {
	case TransitOriginState:
		if err := ship.Refuel(); err != nil {
			ship.log.Error(err.Error())
			return
		}
		err := ship.Transit(m.origin)
		if err != nil {
			ship.log.Error(err.Error())
			ship.Transit("X1-HK42-A1")
			return
		}

		m.state = ExtractState
	case ExtractState:

		if err := ship.Refuel(); err != nil {
			ship.log.Error(err.Error())
			return
		}

		if ship.HasSurveyor() && !m.HasValidSurveys() {
			if surveys, err := ship.Survey(); err == nil {
				m.surveys = surveys
				return // to enable cooldown
			} else {
				ship.log.Warn(errors.Wrap(err, "failed to create survey").Error())
			}
		}

		shitGoods := []string{"ICE_WATER"}
		for _, g := range shitGoods {
			if ship.HasGood("ICE_WATER") {
				if err := ship.Jettison(g); err != nil {
					ship.log.Error(errors.Wrap(err, "Failed to jettison worthloss good: "+g).Error())
				}
			}
		}

		if ship.IsCargoFull() {
			m.state = TransitBestMarket
			return
		}

		if err := ship.Extract(api.Survey{}); err != nil {
			ship.log.Error(err.Error())
		}

		// if m.HasValidSurveys() {
		// 	if err := ship.Extract(m.surveys[0]); err != nil {
		// 		ship.log.Error(err.Error())
		// 	}
		// } else {
		// 	if err := ship.Extract(api.Survey{}); err != nil {
		// 		ship.log.Error(err.Error())
		// 	}
		// }

	case TransitBestMarket:
		wps, err := m.repo.FindMarketsForGoods(ship.InventorySymbols())
		if err != nil {
			ship.log.Error(err.Error())
		}
		if len(wps) == 0 {
			if ship.IsCargoFull() {
				ship.log.Warn("cargo is full with nowhere to sell!")
				return
			}
			ship.log.Debug("no available markets, going to extract")
			m.state = TransitOriginState
			return
		}

		for _, wp := range wps {
			err := ship.Transit(wp)
			if err != nil {
				ship.log.Warn(fmt.Sprintf("couldnt navigate to %s, (maybe low fuel) using next best", wp))
				continue
			}
			m.state = SellState
			return
		}

		ship.log.Warn("couldnt navigate to any waypoints, heading to A1")
		err = ship.Transit("X1-HK42-A1")
		if err != nil {
			ship.log.Error(err.Error())
			return
		}

	case SellState:
		if err := ship.Refuel(); err != nil {
			ship.log.Error(err.Error())
			return
		}
		for _, tg := range ship.state.Cargo.Inventory {
			if err := ship.Sell(string(tg.Symbol)); err != nil {
				ship.log.Error(errors.Wrap(err, "SellState failed to SellAll").Error())
			}
		}

		if ship.IsCargoEmpty() {
			m.state = TransitOriginState
		}

		m.state = TransitBestMarket

		return
	}
}

func (m *MiningMission) ComputeState(ship *Ship) {

	if ship.IsCargoFull() {
		m.state = SellState
		return
	}

	if ship.At(m.origin) {
		m.state = ExtractState
		return
	}

	m.state = TransitOriginState
}
