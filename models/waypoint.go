package models

type Waypoint struct {
	Symbol string `db:"symbol"`
	Type   string `db:"type"`
	X      int    `db:"x"`
	Y      int    `db:"y"`

	// Dist is the distance from the sun, used for ui rendering
	Dist float64
}
