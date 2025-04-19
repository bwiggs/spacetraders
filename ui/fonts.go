package main

import (
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func loadFont(path string, size float64) font.Face {
	tt, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	ttf, err := opentype.Parse(tt)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	return face
}
