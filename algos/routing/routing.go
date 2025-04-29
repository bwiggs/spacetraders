package algos

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/utils"
)

type Edge struct {
	To           string  // Destination node
	Distance     float64 // Distance between nodes
	FuelRequired int     // Distance between nodes
}

type Node struct {
	ID        string
	FuelPrice float64 // Price per unit fuel at this node
	Edges     []Edge
	HasFuel   bool
}

type State struct {
	NodeID        string
	Priority      float64
	FuelRemaining int
}

type PriorityQueue []*State

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	*pq = append(*pq, x.(*State))
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func djikstra(nodes map[string]Node, start, end string, shipFuelCapacity int) (float64, []string) {
	const INF = math.MaxFloat64
	costs := make(map[string]float64)
	prev := make(map[string]string)

	for id := range nodes {
		costs[id] = INF
	}
	costs[start] = 0

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &State{NodeID: start, Priority: 0, FuelRemaining: shipFuelCapacity})

	for pq.Len() > 0 {
		state := heap.Pop(pq).(*State)
		currentID := state.NodeID
		remainingFuel := state.FuelRemaining
		currentNode := nodes[currentID]

		if currentID == end {
			// Reconstruct path
			path := []string{}
			for at := end; at != ""; at = prev[at] {
				path = append([]string{at}, path...)
			}
			return costs[end], path
		}

		for _, edge := range currentNode.Edges {

			// Check if we have enough fuel without refueling
			if edge.FuelRequired <= remainingFuel {
				newRemainingFuel := remainingFuel - edge.FuelRequired
				edgeCost := costs[currentID] + edge.Distance

				if edgeCost < costs[edge.To] {
					costs[edge.To] = edgeCost
					prev[edge.To] = currentID
					heap.Push(pq, &State{NodeID: edge.To, Priority: edgeCost, FuelRemaining: newRemainingFuel})
				}
			} else {
				// try a drift path
				newRemainingFuel := remainingFuel - 1
				driftCost := 10000.0
				edgeCost := costs[currentID] + edge.Distance + driftCost

				if edgeCost < costs[edge.To] {
					costs[edge.To] = edgeCost
					prev[edge.To] = currentID
					heap.Push(pq, &State{NodeID: edge.To, Priority: edgeCost, FuelRemaining: newRemainingFuel})
				}
			}

			// Option 2: Refuel, if possible
			if currentNode.HasFuel && shipFuelCapacity >= edge.FuelRequired {
				// Calculate how much fuel is needed to refill fully
				blocks := math.Ceil(float64(shipFuelCapacity-remainingFuel) / 100.0)
				refuelCost := blocks * currentNode.FuelPrice

				// After refueling, we have full fuel minus the fuel needed for this edge
				newRemainingFuel := shipFuelCapacity - edge.FuelRequired
				edgeCost := costs[currentID] + edge.Distance + refuelCost

				if edgeCost < costs[edge.To] {
					costs[edge.To] = edgeCost
					prev[edge.To] = currentID
					heap.Push(pq, &State{NodeID: edge.To, Priority: edgeCost, FuelRemaining: newRemainingFuel})
				}
			}
		}
	}
	return INF, nil // No path found

}

func FindPath(ship *api.Ship, destSymbol string, waypoints []*models.Waypoint) (float64, []string) {
	nodes := map[string]Node{}

	// Create nodes based on waypoints
	for i, wp := range waypoints {
		n := Node{ID: string(ship.Nav.WaypointSymbol), FuelPrice: 72, Edges: []Edge{}, HasFuel: wp.CanRefuel()}

		for j, neighbor := range waypoints {
			if i != j {
				// Assuming a function to calculate distance between waypoints
				distance := utils.Distance2dInt(wp.X, wp.Y, neighbor.X, neighbor.Y)
				fuelCost, err := calcFuelCost(distance, string(ship.Nav.FlightMode))
				if err != nil {
					panic(err)
				}
				n.Edges = append(n.Edges, Edge{To: neighbor.Symbol, Distance: float64(distance), FuelRequired: fuelCost})
			}
		}
		nodes[wp.Symbol] = n
	}

	cost, path := djikstra(nodes, string(ship.Nav.WaypointSymbol), destSymbol, ship.Fuel.Capacity)
	return cost, path
}

// https://github.com/SpaceTradersAPI/api-docs/wiki/Travel-Fuel-and-Time
func calcFuelCost(distance int, flightMode string) (int, error) {
	switch flightMode {
	case "STEALTH":
		return max(1, distance), nil
	case "CRUISE":
		return max(1, distance), nil
	case "BURN":
		return max(2, distance*2), nil
	case "DRIFT":
		return 1, nil
	}

	return math.MaxInt, fmt.Errorf("unknown flight mode: %s", flightMode)
}
