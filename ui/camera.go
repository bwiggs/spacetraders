package main

import (
	"log/slog"
	"math"
)

// Camera2D manages zoom, pan, and world-to-screen transforms
type Camera2D struct {
	CenterX float64
	CenterY float64
	Zoom    float64
}

func NewCamera2D() *Camera2D {
	return &Camera2D{
		CenterX: 0,
		CenterY: 0,
		Zoom:    1.2,
	}
}

// LookAt sets the camera to center on the given world coordinates
func (c *Camera2D) LookAt(x, y float64) {
	slog.Debug("camera: LookAt", "x", x, "y", y)
	c.CenterX = x
	c.CenterY = y
}

func (c *Camera2D) GetTransform(screenWidth, screenHeight int) (scale, offsetX, offsetY float64) {
	scale = c.Zoom
	offsetX = float64(screenWidth)/2 - c.CenterX*scale
	offsetY = float64(screenHeight)/2 - c.CenterY*scale
	return
}

func (c *Camera2D) WorldToScreen(wx, wy float64, screenWidth, screenHeight int) (float64, float64) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight)
	sx := wx*scale + offsetX
	sy := wy*scale + offsetY
	return sx, sy
}

func (c *Camera2D) ScreenToWorld(sx, sy float64, screenWidth, screenHeight int) (float64, float64) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight)
	wx := (sx - offsetX) / scale
	wy := (sy - offsetY) / scale
	return wx, wy
}

func (c *Camera2D) GetWorldBounds(screenWidth, screenHeight int) (minX, maxX, minY, maxY float32) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight)

	topLeftX, topLeftY := (0-offsetX)/scale, (0-offsetY)/scale
	bottomRightX, bottomRightY := (float64(screenWidth)-offsetX)/scale, (float64(screenHeight)-offsetY)/scale

	minX = float32(math.Min(topLeftX, bottomRightX))
	maxX = float32(math.Max(topLeftX, bottomRightX))
	minY = float32(math.Min(topLeftY, bottomRightY))
	maxY = float32(math.Max(topLeftY, bottomRightY))

	return
}
