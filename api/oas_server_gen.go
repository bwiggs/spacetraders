// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
)

// Handler handles operations described by OpenAPI v3 specification.
type Handler interface {
	// AcceptContract implements accept-contract operation.
	//
	// Accept a contract by ID.
	// You can only accept contracts that were offered to you, were not accepted yet, and whose deadlines
	// has not passed yet.
	//
	// POST /my/contracts/{contractId}/accept
	AcceptContract(ctx context.Context, params AcceptContractParams) (*AcceptContractOK, error)
	// CreateChart implements create-chart operation.
	//
	// Command a ship to chart the waypoint at its current location.
	// Most waypoints in the universe are uncharted by default. These waypoints have their traits hidden
	// until they have been charted by a ship.
	// Charting a waypoint will record your agent as the one who created the chart, and all other agents
	// would also be able to see the waypoint's traits.
	//
	// POST /my/ships/{shipSymbol}/chart
	CreateChart(ctx context.Context, params CreateChartParams) (*CreateChartCreated, error)
	// CreateShipShipScan implements create-ship-ship-scan operation.
	//
	// Scan for nearby ships, retrieving information for all ships in range.
	// Requires a ship to have the `Sensor Array` mount installed to use.
	// The ship will enter a cooldown after using this function, during which it cannot execute certain
	// actions.
	//
	// POST /my/ships/{shipSymbol}/scan/ships
	CreateShipShipScan(ctx context.Context, params CreateShipShipScanParams) (*CreateShipShipScanCreated, error)
	// CreateShipSystemScan implements create-ship-system-scan operation.
	//
	// Scan for nearby systems, retrieving information on the systems' distance from the ship and their
	// waypoints. Requires a ship to have the `Sensor Array` mount installed to use.
	// The ship will enter a cooldown after using this function, during which it cannot execute certain
	// actions.
	//
	// POST /my/ships/{shipSymbol}/scan/systems
	CreateShipSystemScan(ctx context.Context, params CreateShipSystemScanParams) (*CreateShipSystemScanCreated, error)
	// CreateShipWaypointScan implements create-ship-waypoint-scan operation.
	//
	// Scan for nearby waypoints, retrieving detailed information on each waypoint in range. Scanning
	// uncharted waypoints will allow you to ignore their uncharted state and will list the waypoints'
	// traits.
	// Requires a ship to have the `Sensor Array` mount installed to use.
	// The ship will enter a cooldown after using this function, during which it cannot execute certain
	// actions.
	//
	// POST /my/ships/{shipSymbol}/scan/waypoints
	CreateShipWaypointScan(ctx context.Context, params CreateShipWaypointScanParams) (*CreateShipWaypointScanCreated, error)
	// CreateSurvey implements create-survey operation.
	//
	// Create surveys on a waypoint that can be extracted such as asteroid fields. A survey focuses on
	// specific types of deposits from the extracted location. When ships extract using this survey, they
	// are guaranteed to procure a high amount of one of the goods in the survey.
	// In order to use a survey, send the entire survey details in the body of the extract request.
	// Each survey may have multiple deposits, and if a symbol shows up more than once, that indicates a
	// higher chance of extracting that resource.
	// Your ship will enter a cooldown after surveying in which it is unable to perform certain actions.
	// Surveys will eventually expire after a period of time or will be exhausted after being extracted
	// several times based on the survey's size. Multiple ships can use the same survey for extraction.
	// A ship must have the `Surveyor` mount installed in order to use this function.
	//
	// POST /my/ships/{shipSymbol}/survey
	CreateSurvey(ctx context.Context, params CreateSurveyParams) (*CreateSurveyCreated, error)
	// DeliverContract implements deliver-contract operation.
	//
	// Deliver cargo to a contract.
	// In order to use this API, a ship must be at the delivery location (denoted in the delivery terms
	// as `destinationSymbol` of a contract) and must have a number of units of a good required by this
	// contract in its cargo.
	// Cargo that was delivered will be removed from the ship's cargo.
	//
	// POST /my/contracts/{contractId}/deliver
	DeliverContract(ctx context.Context, req OptDeliverContractReq, params DeliverContractParams) (*DeliverContractOK, error)
	// DockShip implements dock-ship operation.
	//
	// Attempt to dock your ship at its current location. Docking will only succeed if your ship is
	// capable of docking at the time of the request.
	// Docked ships can access elements in their current location, such as the market or a shipyard, but
	// cannot do actions that require the ship to be above surface such as navigating or extracting.
	// The endpoint is idempotent - successive calls will succeed even if the ship is already docked.
	//
	// POST /my/ships/{shipSymbol}/dock
	DockShip(ctx context.Context, params DockShipParams) (*DockShipOK, error)
	// ExtractResources implements extract-resources operation.
	//
	// Extract resources from a waypoint that can be extracted, such as asteroid fields, into your ship.
	// Send an optional survey as the payload to target specific yields.
	// The ship must be in orbit to be able to extract and must have mining equipments installed that can
	// extract goods, such as the `Gas Siphon` mount for gas-based goods or `Mining Laser` mount for
	// ore-based goods.
	// The survey property is now deprecated. See the `extract/survey` endpoint for more details.
	//
	// POST /my/ships/{shipSymbol}/extract
	ExtractResources(ctx context.Context, req OptExtractResourcesReq, params ExtractResourcesParams) (*ExtractResourcesCreated, error)
	// ExtractResourcesWithSurvey implements extract-resources-with-survey operation.
	//
	// Use a survey when extracting resources from a waypoint. This endpoint requires a survey as the
	// payload, which allows your ship to extract specific yields.
	// Send the full survey object as the payload which will be validated according to the signature. If
	// the signature is invalid, or any properties of the survey are changed, the request will fail.
	//
	// POST /my/ships/{shipSymbol}/extract/survey
	ExtractResourcesWithSurvey(ctx context.Context, req OptSurvey, params ExtractResourcesWithSurveyParams) (*ExtractResourcesWithSurveyCreated, error)
	// FulfillContract implements fulfill-contract operation.
	//
	// Fulfill a contract. Can only be used on contracts that have all of their delivery terms fulfilled.
	//
	// POST /my/contracts/{contractId}/fulfill
	FulfillContract(ctx context.Context, params FulfillContractParams) (*FulfillContractOK, error)
	// GetAgent implements get-agent operation.
	//
	// Fetch agent details.
	//
	// GET /agents/{agentSymbol}
	GetAgent(ctx context.Context, params GetAgentParams) (*GetAgentOK, error)
	// GetAgents implements get-agents operation.
	//
	// Fetch agents details.
	//
	// GET /agents
	GetAgents(ctx context.Context, params GetAgentsParams) (*GetAgentsOK, error)
	// GetConstruction implements get-construction operation.
	//
	// Get construction details for a waypoint. Requires a waypoint with a property of
	// `isUnderConstruction` to be true.
	//
	// GET /systems/{systemSymbol}/waypoints/{waypointSymbol}/construction
	GetConstruction(ctx context.Context, params GetConstructionParams) (*GetConstructionOK, error)
	// GetContract implements get-contract operation.
	//
	// Get the details of a contract by ID.
	//
	// GET /my/contracts/{contractId}
	GetContract(ctx context.Context, params GetContractParams) (*GetContractOK, error)
	// GetContracts implements get-contracts operation.
	//
	// Return a paginated list of all your contracts.
	//
	// GET /my/contracts
	GetContracts(ctx context.Context, params GetContractsParams) (*GetContractsOK, error)
	// GetFaction implements get-faction operation.
	//
	// View the details of a faction.
	//
	// GET /factions/{factionSymbol}
	GetFaction(ctx context.Context, params GetFactionParams) (*GetFactionOK, error)
	// GetFactions implements get-factions operation.
	//
	// Return a paginated list of all the factions in the game.
	//
	// GET /factions
	GetFactions(ctx context.Context, params GetFactionsParams) (*GetFactionsOK, error)
	// GetJumpGate implements get-jump-gate operation.
	//
	// Get jump gate details for a waypoint. Requires a waypoint of type `JUMP_GATE` to use.
	// Waypoints connected to this jump gate can be.
	//
	// GET /systems/{systemSymbol}/waypoints/{waypointSymbol}/jump-gate
	GetJumpGate(ctx context.Context, params GetJumpGateParams) (*GetJumpGateOK, error)
	// GetMarket implements get-market operation.
	//
	// Retrieve imports, exports and exchange data from a marketplace. Requires a waypoint that has the
	// `Marketplace` trait to use.
	// Send a ship to the waypoint to access trade good prices and recent transactions. Refer to the
	// [Market Overview page](https://docs.spacetraders.io/game-concepts/markets) to gain better a
	// understanding of the market in the game.
	//
	// GET /systems/{systemSymbol}/waypoints/{waypointSymbol}/market
	GetMarket(ctx context.Context, params GetMarketParams) (*GetMarketOK, error)
	// GetMounts implements get-mounts operation.
	//
	// Get the mounts installed on a ship.
	//
	// GET /my/ships/{shipSymbol}/mounts
	GetMounts(ctx context.Context, params GetMountsParams) (*GetMountsOK, error)
	// GetMyAgent implements get-my-agent operation.
	//
	// Fetch your agent's details.
	//
	// GET /my/agent
	GetMyAgent(ctx context.Context) (*GetMyAgentOK, error)
	// GetMyShip implements get-my-ship operation.
	//
	// Retrieve the details of a ship under your agent's ownership.
	//
	// GET /my/ships/{shipSymbol}
	GetMyShip(ctx context.Context, params GetMyShipParams) (*GetMyShipOK, error)
	// GetMyShipCargo implements get-my-ship-cargo operation.
	//
	// Retrieve the cargo of a ship under your agent's ownership.
	//
	// GET /my/ships/{shipSymbol}/cargo
	GetMyShipCargo(ctx context.Context, params GetMyShipCargoParams) (*GetMyShipCargoOK, error)
	// GetMyShips implements get-my-ships operation.
	//
	// Return a paginated list of all of ships under your agent's ownership.
	//
	// GET /my/ships
	GetMyShips(ctx context.Context, params GetMyShipsParams) (*GetMyShipsOK, error)
	// GetShipCooldown implements get-ship-cooldown operation.
	//
	// Retrieve the details of your ship's reactor cooldown. Some actions such as activating your jump
	// drive, scanning, or extracting resources taxes your reactor and results in a cooldown.
	// Your ship cannot perform additional actions until your cooldown has expired. The duration of your
	// cooldown is relative to the power consumption of the related modules or mounts for the action
	// taken.
	// Response returns a 204 status code (no-content) when the ship has no cooldown.
	//
	// GET /my/ships/{shipSymbol}/cooldown
	GetShipCooldown(ctx context.Context, params GetShipCooldownParams) (GetShipCooldownRes, error)
	// GetShipNav implements get-ship-nav operation.
	//
	// Get the current nav status of a ship.
	//
	// GET /my/ships/{shipSymbol}/nav
	GetShipNav(ctx context.Context, params GetShipNavParams) (*GetShipNavOK, error)
	// GetShipyard implements get-shipyard operation.
	//
	// Get the shipyard for a waypoint. Requires a waypoint that has the `Shipyard` trait to use. Send a
	// ship to the waypoint to access data on ships that are currently available for purchase and recent
	// transactions.
	//
	// GET /systems/{systemSymbol}/waypoints/{waypointSymbol}/shipyard
	GetShipyard(ctx context.Context, params GetShipyardParams) (*GetShipyardOK, error)
	// GetStatus implements get-status operation.
	//
	// Return the status of the game server.
	// This also includes a few global elements, such as announcements, server reset dates and
	// leaderboards.
	//
	// GET /
	GetStatus(ctx context.Context) (*GetStatusOK, error)
	// GetSystem implements get-system operation.
	//
	// Get the details of a system.
	//
	// GET /systems/{systemSymbol}
	GetSystem(ctx context.Context, params GetSystemParams) (*GetSystemOK, error)
	// GetSystemWaypoints implements get-system-waypoints operation.
	//
	// Return a paginated list of all of the waypoints for a given system.
	// If a waypoint is uncharted, it will return the `Uncharted` trait instead of its actual traits.
	//
	// GET /systems/{systemSymbol}/waypoints
	GetSystemWaypoints(ctx context.Context, params GetSystemWaypointsParams) (*GetSystemWaypointsOK, error)
	// GetSystems implements get-systems operation.
	//
	// Return a paginated list of all systems.
	//
	// GET /systems
	GetSystems(ctx context.Context, params GetSystemsParams) (*GetSystemsOK, error)
	// GetWaypoint implements get-waypoint operation.
	//
	// View the details of a waypoint.
	// If the waypoint is uncharted, it will return the 'Uncharted' trait instead of its actual traits.
	//
	// GET /systems/{systemSymbol}/waypoints/{waypointSymbol}
	GetWaypoint(ctx context.Context, params GetWaypointParams) (*GetWaypointOK, error)
	// InstallMount implements install-mount operation.
	//
	// Install a mount on a ship.
	// In order to install a mount, the ship must be docked and located in a waypoint that has a
	// `Shipyard` trait. The ship also must have the mount to install in its cargo hold.
	// An installation fee will be deduced by the Shipyard for installing the mount on the ship.
	//
	// POST /my/ships/{shipSymbol}/mounts/install
	InstallMount(ctx context.Context, req OptInstallMountReq, params InstallMountParams) (*InstallMountCreated, error)
	// Jettison implements jettison operation.
	//
	// Jettison cargo from your ship's cargo hold.
	//
	// POST /my/ships/{shipSymbol}/jettison
	Jettison(ctx context.Context, req OptJettisonReq, params JettisonParams) (*JettisonOK, error)
	// JumpShip implements jump-ship operation.
	//
	// Jump your ship instantly to a target connected waypoint. The ship must be in orbit to execute a
	// jump.
	// A unit of antimatter is purchased and consumed from the market when jumping. The price of
	// antimatter is determined by the market and is subject to change. A ship can only jump to connected
	// waypoints.
	//
	// POST /my/ships/{shipSymbol}/jump
	JumpShip(ctx context.Context, req OptJumpShipReq, params JumpShipParams) (*JumpShipOK, error)
	// NavigateShip implements navigate-ship operation.
	//
	// Navigate to a target destination. The ship must be in orbit to use this function. The destination
	// waypoint must be within the same system as the ship's current location. Navigating will consume
	// the necessary fuel from the ship's manifest based on the distance to the target waypoint.
	// The returned response will detail the route information including the expected time of arrival.
	// Most ship actions are unavailable until the ship has arrived at it's destination.
	// To travel between systems, see the ship's Warp or Jump actions.
	//
	// POST /my/ships/{shipSymbol}/navigate
	NavigateShip(ctx context.Context, req OptNavigateShipReq, params NavigateShipParams) (*NavigateShipOK, error)
	// NegotiateContract implements negotiateContract operation.
	//
	// Negotiate a new contract with the HQ.
	// In order to negotiate a new contract, an agent must not have ongoing or offered contracts over the
	// allowed maximum amount. Currently the maximum contracts an agent can have at a time is 1.
	// Once a contract is negotiated, it is added to the list of contracts offered to the agent, which
	// the agent can then accept.
	// The ship must be present at any waypoint with a faction present to negotiate a contract with that
	// faction.
	//
	// POST /my/ships/{shipSymbol}/negotiate/contract
	NegotiateContract(ctx context.Context, params NegotiateContractParams) (*NegotiateContractCreated, error)
	// OrbitShip implements orbit-ship operation.
	//
	// Attempt to move your ship into orbit at its current location. The request will only succeed if
	// your ship is capable of moving into orbit at the time of the request.
	// Orbiting ships are able to do actions that require the ship to be above surface such as navigating
	// or extracting, but cannot access elements in their current waypoint, such as the market or a
	// shipyard.
	// The endpoint is idempotent - successive calls will succeed even if the ship is already in orbit.
	//
	// POST /my/ships/{shipSymbol}/orbit
	OrbitShip(ctx context.Context, params OrbitShipParams) (*OrbitShipOK, error)
	// PatchShipNav implements patch-ship-nav operation.
	//
	// Update the nav configuration of a ship.
	// Currently only supports configuring the Flight Mode of the ship, which affects its speed and fuel
	// consumption.
	//
	// PATCH /my/ships/{shipSymbol}/nav
	PatchShipNav(ctx context.Context, req OptPatchShipNavReq, params PatchShipNavParams) (*PatchShipNavOK, error)
	// PurchaseCargo implements purchase-cargo operation.
	//
	// Purchase cargo from a market.
	// The ship must be docked in a waypoint that has `Marketplace` trait, and the market must be selling
	// a good to be able to purchase it.
	// The maximum amount of units of a good that can be purchased in each transaction are denoted by the
	// `tradeVolume` value of the good, which can be viewed by using the Get Market action.
	// Purchased goods are added to the ship's cargo hold.
	//
	// POST /my/ships/{shipSymbol}/purchase
	PurchaseCargo(ctx context.Context, req OptPurchaseCargoReq, params PurchaseCargoParams) (*PurchaseCargoCreated, error)
	// PurchaseShip implements purchase-ship operation.
	//
	// Purchase a ship from a Shipyard. In order to use this function, a ship under your agent's
	// ownership must be in a waypoint that has the `Shipyard` trait, and the Shipyard must sell the type
	// of the desired ship.
	// Shipyards typically offer ship types, which are predefined templates of ships that have dedicated
	// roles. A template comes with a preset of an engine, a reactor, and a frame. It may also include a
	// few modules and mounts.
	//
	// POST /my/ships
	PurchaseShip(ctx context.Context, req OptPurchaseShipReq) (*PurchaseShipCreated, error)
	// RefuelShip implements refuel-ship operation.
	//
	// Refuel your ship by buying fuel from the local market.
	// Requires the ship to be docked in a waypoint that has the `Marketplace` trait, and the market must
	// be selling fuel in order to refuel.
	// Each fuel bought from the market replenishes 100 units in your ship's fuel.
	// Ships will always be refuel to their frame's maximum fuel capacity when using this action.
	//
	// POST /my/ships/{shipSymbol}/refuel
	RefuelShip(ctx context.Context, req OptRefuelShipReq, params RefuelShipParams) (*RefuelShipOK, error)
	// Register implements register operation.
	//
	// Creates a new agent and ties it to an account.
	// The agent symbol must consist of a 3-14 character string, and will be used to represent your agent.
	//  This symbol will prefix the symbol of every ship you own. Agent symbols will be cast to all
	// uppercase characters.
	// This new agent will be tied to a starting faction of your choice, which determines your starting
	// location, and will be granted an authorization token, a contract with their starting faction, a
	// command ship that can fly across space with advanced capabilities, a small probe ship that can be
	// used for reconnaissance, and 150,000 credits.
	// > #### Keep your token safe and secure
	// >
	// > Save your token during the alpha phase. There is no way to regenerate this token without
	// starting a new agent. In the future you will be able to generate and manage your tokens from the
	// SpaceTraders website.
	// If you are new to SpaceTraders, It is recommended to register with the COSMIC faction, a faction
	// that is well connected to the rest of the universe. After registering, you should try our
	// interactive [quickstart guide](https://docs.spacetraders.io/quickstart/new-game) which will walk
	// you through basic API requests in just a few minutes.
	//
	// POST /register
	Register(ctx context.Context, req OptRegisterReq) (*RegisterCreated, error)
	// RemoveMount implements remove-mount operation.
	//
	// Remove a mount from a ship.
	// The ship must be docked in a waypoint that has the `Shipyard` trait, and must have the desired
	// mount that it wish to remove installed.
	// A removal fee will be deduced from the agent by the Shipyard.
	//
	// POST /my/ships/{shipSymbol}/mounts/remove
	RemoveMount(ctx context.Context, req OptRemoveMountReq, params RemoveMountParams) (*RemoveMountCreated, error)
	// SellCargo implements sell-cargo operation.
	//
	// Sell cargo in your ship to a market that trades this cargo. The ship must be docked in a waypoint
	// that has the `Marketplace` trait in order to use this function.
	//
	// POST /my/ships/{shipSymbol}/sell
	SellCargo(ctx context.Context, req OptSellCargoReq, params SellCargoParams) (*SellCargoCreated, error)
	// ShipRefine implements ship-refine operation.
	//
	// Attempt to refine the raw materials on your ship. The request will only succeed if your ship is
	// capable of refining at the time of the request. In order to be able to refine, a ship must have
	// goods that can be refined and have installed a `Refinery` module that can refine it.
	// When refining, 30 basic goods will be converted into 10 processed goods.
	//
	// POST /my/ships/{shipSymbol}/refine
	ShipRefine(ctx context.Context, req OptShipRefineReq, params ShipRefineParams) (*ShipRefineCreated, error)
	// SiphonResources implements siphon-resources operation.
	//
	// Siphon gases, such as hydrocarbon, from gas giants.
	// The ship must be in orbit to be able to siphon and must have siphon mounts and a gas processor
	// installed.
	//
	// POST /my/ships/{shipSymbol}/siphon
	SiphonResources(ctx context.Context, params SiphonResourcesParams) (*SiphonResourcesCreated, error)
	// SupplyConstruction implements supply-construction operation.
	//
	// Supply a construction site with the specified good. Requires a waypoint with a property of
	// `isUnderConstruction` to be true.
	// The good must be in your ship's cargo. The good will be removed from your ship's cargo and added
	// to the construction site's materials.
	//
	// POST /systems/{systemSymbol}/waypoints/{waypointSymbol}/construction/supply
	SupplyConstruction(ctx context.Context, req OptSupplyConstructionReq, params SupplyConstructionParams) (*SupplyConstructionCreated, error)
	// TransferCargo implements transfer-cargo operation.
	//
	// Transfer cargo between ships.
	// The receiving ship must be in the same waypoint as the transferring ship, and it must able to hold
	// the additional cargo after the transfer is complete. Both ships also must be in the same state,
	// either both are docked or both are orbiting.
	// The response body's cargo shows the cargo of the transferring ship after the transfer is complete.
	//
	// POST /my/ships/{shipSymbol}/transfer
	TransferCargo(ctx context.Context, req OptTransferCargoReq, params TransferCargoParams) (*TransferCargoOK, error)
	// WarpShip implements warp-ship operation.
	//
	// Warp your ship to a target destination in another system. The ship must be in orbit to use this
	// function and must have the `Warp Drive` module installed. Warping will consume the necessary fuel
	// from the ship's manifest.
	// The returned response will detail the route information including the expected time of arrival.
	// Most ship actions are unavailable until the ship has arrived at its destination.
	//
	// POST /my/ships/{shipSymbol}/warp
	WarpShip(ctx context.Context, req OptWarpShipReq, params WarpShipParams) (*WarpShipOK, error)
}

// Server implements http server based on OpenAPI v3 specification and
// calls Handler to handle requests.
type Server struct {
	h   Handler
	sec SecurityHandler
	baseServer
}

// NewServer creates new Server.
func NewServer(h Handler, sec SecurityHandler, opts ...ServerOption) (*Server, error) {
	s, err := newServerConfig(opts...).baseServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		h:          h,
		sec:        sec,
		baseServer: s,
	}, nil
}