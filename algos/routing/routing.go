package algos

import (
	"container/heap"
	"math"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/models"
	"github.com/bwiggs/spacetraders-go/utils"
)

type Edge struct {
	To       string  // Destination node
	Distance float64 // Distance between nodes
}

type Node struct {
	ID       string
	FuelCost float64 // Price per unit fuel at this node
	Edges    []Edge
}

type State struct {
	NodeID   string
	Distance float64
	Priority float64
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

func djikstra(nodes map[string]Node, start, end string) (float64, []string) {
	const INF = math.MaxFloat64
	distances := make(map[string]float64)
	prev := make(map[string]string)

	for id := range nodes {
		distances[id] = INF
	}
	distances[start] = 0

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &State{NodeID: start, Distance: 0, Priority: 0})

	for pq.Len() > 0 {
		state := heap.Pop(pq).(*State)
		currentID := state.NodeID

		if currentID == end {
			// Reconstruct path
			path := []string{}
			for at := end; at != ""; at = prev[at] {
				path = append([]string{at}, path...)
			}
			return distances[end], path
		}

		currentNode := nodes[currentID]
		for _, edge := range currentNode.Edges {
			newDist := distances[currentID] + edge.Distance
			if newDist < distances[edge.To] {
				distances[edge.To] = newDist
				prev[edge.To] = currentID
				heap.Push(pq, &State{NodeID: edge.To, Distance: newDist, Priority: newDist})
			}
		}
	}
	return INF, nil // No path found

}

func FindPath(ship *api.Ship, destSymbol string, waypoints []models.Waypoint) []string {
	nodes := map[string]Node{}

	// Create nodes based on waypoints
	for i, wp := range waypoints {
		n := Node{ID: string(ship.Nav.WaypointSymbol), FuelCost: 1, Edges: []Edge{}}
		for j, neighbor := range waypoints {
			if i != j {
				// Assuming a function to calculate distance between waypoints

				distance := utils.Distance2dInt(wp.X, wp.Y, neighbor.X, neighbor.Y) // Replace with actual distance calculation
				if distance < 275 {
					n.Edges = append(n.Edges, Edge{To: neighbor.Symbol, Distance: float64(distance)})
				}
			}
		}
		nodes[wp.Symbol] = n
	}

	_, path := djikstra(nodes, string(ship.Nav.WaypointSymbol), destSymbol)
	return path[1:]
}
