package map_art

import (
	_ "embed"
	"image/color"
)

const (
	HeightModeHigher = iota
	HeightModeMiddle
	HeightModeLower
)

//go:embed colors.json
var colorJsonBytes []byte

var (
	availableColors     []MapBlockColor
	colorToMapBlockMode map[color.RGBA]MapBlockMode
	mapArtColorMapping  [][3]uint8
)

// MapBlockColor ..
type MapBlockColor struct {
	HigherColor color.RGBA
	MiddleColor color.RGBA
	LowerColor  color.RGBA
	BlockName   string
}

// MapBlockMode ..
type MapBlockMode struct {
	BlockName  string
	HeightMode uint8
}

// SingleMapBlock ..
type SingleMapBlock struct {
	BlockMode MapBlockMode
	PosY      int16
}
