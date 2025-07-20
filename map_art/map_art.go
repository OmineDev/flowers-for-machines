package map_art

import (
	"encoding/json"
	"image/color"
)

func init() {
	var colors [][]any

	err := json.Unmarshal(colorJsonBytes, &colors)
	if err != nil {
		panic(err)
	}

	availableColors = make([]MapBlockColor, len(colors))
	colorToMapBlockMode = make(map[color.RGBA]MapBlockMode)

	for index, value := range colors {
		blockName := value[3].(string)

		higherColorList := value[0].([]any)
		higherColor := color.RGBA{
			uint8(higherColorList[0].(float64)),
			uint8(higherColorList[1].(float64)),
			uint8(higherColorList[2].(float64)),
			uint8(higherColorList[3].(float64)),
		}
		colorToMapBlockMode[higherColor] = MapBlockMode{
			BlockName:  blockName,
			HeightMode: HeightModeHigher,
		}

		middleColorList := value[1].([]any)
		middleColor := color.RGBA{
			uint8(middleColorList[0].(float64)),
			uint8(middleColorList[1].(float64)),
			uint8(middleColorList[2].(float64)),
			uint8(middleColorList[3].(float64)),
		}
		colorToMapBlockMode[middleColor] = MapBlockMode{
			BlockName:  blockName,
			HeightMode: HeightModeMiddle,
		}

		lowerColorList := value[2].([]any)
		lowerColor := color.RGBA{
			uint8(lowerColorList[0].(float64)),
			uint8(lowerColorList[1].(float64)),
			uint8(lowerColorList[2].(float64)),
			uint8(lowerColorList[3].(float64)),
		}
		colorToMapBlockMode[lowerColor] = MapBlockMode{
			BlockName:  blockName,
			HeightMode: HeightModeLower,
		}

		availableColors[index] = MapBlockColor{
			HigherColor: higherColor,
			MiddleColor: middleColor,
			LowerColor:  lowerColor,
			BlockName:   blockName,
		}
	}

	for _, value := range availableColors {
		mapArtColorMapping = append(mapArtColorMapping, [3]uint8{
			value.HigherColor.R,
			value.HigherColor.G,
			value.HigherColor.B,
		})
		mapArtColorMapping = append(mapArtColorMapping, [3]uint8{
			value.MiddleColor.R,
			value.MiddleColor.G,
			value.MiddleColor.B,
		})
		mapArtColorMapping = append(mapArtColorMapping, [3]uint8{
			value.LowerColor.R,
			value.LowerColor.G,
			value.LowerColor.B,
		})
	}
}
