package actors

type State string

const (
	TransitDestState     = "TRANSIT_DEST"
	TransitOriginState   = "TRANSIT_ORIGIN"
	BuyState             = "BUY"
	SellState            = "SELL"
	IdleState            = "IDLE"
	ErrorState           = "ERROR"
	ExtractState         = "EXTRACT"
	StuckState           = "STUCK"
	ContractDeliverState = "CONTRACT_DELIVER"
	TransitBestMarket    = "TRANSIT_BEST_MARKET"
)
