package main

import (
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
	c.CenterX = x
	c.CenterY = y
}

// GetTransform returns the world->screen scale and offset
func (c *Camera2D) GetTransform(screenWidth, screenHeight int, worldSize float64) (scale, offsetX, offsetY float64) {
	sw := float64(screenWidth)
	sh := float64(screenHeight)

	// Base scale on the smaller dimension to maintain aspect ratio
	baseScale := math.Min(sw/worldSize, sh/worldSize)

	scale = baseScale * c.Zoom

	// Offset so CenterX/Y is in center of screen
	offsetX = (sw / 2) - (c.CenterX * scale)
	offsetY = (sh / 2) - (c.CenterY * scale)

	return
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera2D) WorldToScreen(wx, wy float64, screenWidth, screenHeight int, worldSize float64) (float32, float32) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight, worldSize)
	sx := wx*scale + offsetX
	sy := wy*scale + offsetY
	return float32(sx), float32(sy)
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera2D) ScreenToWorld(sx, sy float64, screenWidth, screenHeight int, worldSize float64) (float64, float64) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight, worldSize)
	wx := (sx - offsetX) / scale
	wy := (sy - offsetY) / scale
	return wx, wy
}

func (c *Camera2D) GetWorldBounds(screenWidth, screenHeight int, worldSize float64) (minX, maxX, minY, maxY float32) {
	scale, offsetX, offsetY := c.GetTransform(screenWidth, screenHeight, worldSize)

	topLeftX, topLeftY := (0-offsetX)/scale, (0-offsetY)/scale
	bottomRightX, bottomRightY := (float64(screenWidth)-offsetX)/scale, (float64(screenHeight)-offsetY)/scale

	minX = float32(math.Min(topLeftX, bottomRightX))
	maxX = float32(math.Max(topLeftX, bottomRightX))
	minY = float32(math.Min(topLeftY, bottomRightY))
	maxY = float32(math.Max(topLeftY, bottomRightY))

	return
}
