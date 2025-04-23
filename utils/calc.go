package utils

import "math"

func Distance2dInt(x1, y1, x2, y2 int) int {
	fx1 := float64(x1)
	fy1 := float64(y1)
	fx2 := float64(x2)
	fy2 := float64(y2)
	return int(math.Sqrt(math.Pow(fx1-fx2, 2) + math.Pow(fy1-fy2, 2)))
}
