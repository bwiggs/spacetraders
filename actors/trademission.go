package actors

import (
	"fmt"
	"time"
)

type TradeMission struct {
	good   string
	origin string
	dest   string
	state  State
}

func NewTradeMission(good, origin, dest string) *TradeMission {
	return &TradeMission{
		good:   good,
		origin: origin,
		dest:   dest,
		state:  IdleState,
	}
}

func (m *TradeMission) Execute(ship *Ship) {
	ship.log.Debug("TradeMission Executing", "state", m.state)

	if m.state == IdleState {
		m.ComputeState(ship)
	}

	switch m.state {
	case TransitOriginState:
		if err := ship.Refuel(); err != nil {
			ship.log.Error(err.Error())
			return
		}
		if err := ship.Transit(m.origin); err != nil {
			ship.log.Error(err.Error())
			return
		}
		m.state = BuyState
	case BuyState:
		if err := ship.Buy(m.good, m.origin); err != nil {
			ship.log.Error(err.Error())
			return
		}
		m.state = TransitDestState
	case TransitDestState:
		if err := ship.Refuel(); err != nil {
			ship.log.Error(err.Error())
			return
		}
		if err := ship.Transit(m.dest); err != nil {
			ship.log.Error(err.Error())
			return
		}
		m.state = SellState
	case SellState:
		if err := ship.Sell(m.good, m.dest); err != nil {
			ship.log.Error(err.Error())
			return
		}
		m.state = TransitOriginState
	}
}

func (m *TradeMission) ComputeState(ship *Ship) {
	cooldown := time.Until(ship.state.Cooldown.Expiration.Value)
	if cooldown > 0 {
		ship.log.Info(fmt.Sprintf("cooldown: %s second", cooldown))
		time.Sleep(cooldown)
	}

	arrival := time.Until(ship.state.Nav.Route.Arrival)
	if arrival > 0 {
		ship.log.Info(fmt.Sprintf("transiting: %s second", arrival))
		time.Sleep(arrival)
	}

	if ship.At(m.dest) {
		if ship.HasGood(m.good) {
			m.state = SellState
			return
		}
		m.state = TransitOriginState
		return
	}

	if ship.At(m.origin) {
		if !ship.HasGood(m.good) {
			m.state = BuyState
			return
		}
		m.state = TransitDestState
		return
	}

}
