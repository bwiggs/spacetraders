package main

import (
	"fmt"
	"time"
)

func Clamp[T ~int | ~float32 | ~float64](val, min, max T) T {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	var out string
	if h > 0 {
		out += fmt.Sprintf("%dh ", h)
	}
	if m > 0 || h > 0 {
		out += fmt.Sprintf("%dm ", m)
	}
	out += fmt.Sprintf("%ds", s)

	return out
}

func Interpolate(origin, destination Point2D, startTime, endTime, currentTime time.Time) Point2D {
	totalDuration := endTime.Sub(startTime).Seconds()
	elapsed := currentTime.Sub(startTime).Seconds()

	progress := elapsed / totalDuration
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}

	return Point2D{
		X: origin.X + (destination.X-origin.X)*progress,
		Y: origin.Y + (destination.Y-origin.Y)*progress,
	}
}
