package actors

type Mission interface {
	Execute(*Ship)
}

type IdleMission struct {
}

func NewIdleMission() *IdleMission {
	return &IdleMission{}
}

func (m *IdleMission) Execute(ship *Ship) {
	ship.log.Info("sleeping")
}
