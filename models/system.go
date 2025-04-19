package models

type System struct {
	Symbol        string `db:"symbol"`
	Constellation string `db:"constellation"`
	Name          string `db:"name"`
	Type          string `db:"type"`
	SectorSymbol  string `db:"sector_symbol"`
	X             int    `db:"x"`
	Y             int    `db:"y"`

	// Dist is the distance from the sun, used for ui rendering
	Dist float64
}
