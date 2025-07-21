package map_art

import (
	"image/color"

	"github.com/OmineDev/flowers-for-machines/utils"
)

// GenerateMapArtStructure ..
func GenerateMapArtStructure(basePos [3]int32, pixels [128][128]color.RGBA) (
	importPos [3]int32,
	blockMatrix [128][129]SingleMapBlock,
) {
	for x := range 128 {
		for z := range 128 {
			pixel := pixels[x][z]

			bestColor := utils.SearchForBestColor(
				[3]uint8{pixel.R, pixel.G, pixel.B},
				mapArtColorMapping,
			)
			bestColorRGBA := color.RGBA{
				bestColor[0],
				bestColor[1],
				bestColor[2],
				255,
			}

			mapBlockMode := colorToMapBlockMode[bestColorRGBA]
			blockMatrix[x][z+1] = SingleMapBlock{BlockMode: mapBlockMode}
		}
	}

	for x := range 128 {
		flag := (x%2 == 0)

		switch blockMatrix[x][1].BlockMode.HeightMode {
		case HeightModeHigher:
			if flag {
				blockMatrix[x][0] = SingleMapBlock{
					BlockMode: MapBlockMode{
						BlockName:  "minecraft:emerald_block",
						HeightMode: HeightModeHigher,
					},
					PosY: -2,
				}
				blockMatrix[x][1].PosY = 0
			} else {
				blockMatrix[x][0] = SingleMapBlock{
					BlockMode: MapBlockMode{
						BlockName:  "minecraft:emerald_block",
						HeightMode: HeightModeHigher,
					},
					PosY: -1,
				}
				blockMatrix[x][1].PosY = 0
			}
		case HeightModeMiddle:
			blockMatrix[x][0] = SingleMapBlock{
				BlockMode: MapBlockMode{
					BlockName:  "minecraft:emerald_block",
					HeightMode: HeightModeMiddle,
				},
				PosY: 0,
			}
			blockMatrix[x][1].PosY = 0
		case HeightModeLower:
			if flag {
				blockMatrix[x][0] = SingleMapBlock{
					BlockMode: MapBlockMode{
						BlockName:  "minecraft:emerald_block",
						HeightMode: HeightModeLower,
					},
					PosY: 1,
				}
				blockMatrix[x][1].PosY = 0
			} else {
				blockMatrix[x][0] = SingleMapBlock{
					BlockMode: MapBlockMode{
						BlockName:  "minecraft:emerald_block",
						HeightMode: HeightModeLower,
					},
					PosY: 2,
				}
				blockMatrix[x][1].PosY = 0
			}
		}
	}

	for x := range 128 {
		flagX := (x%2 == 0)
		for z := 2; z < 129; z++ {
			flagZ := (z%2 == 0)

			lastHeightMode := blockMatrix[x][z-1].BlockMode.HeightMode
			currentHeightMode := blockMatrix[x][z].BlockMode.HeightMode

			lastOnePosY := blockMatrix[x][z-1].PosY
			lastOneOfLastOnePosY := blockMatrix[x][z-2].PosY

			switch lastHeightMode {
			case HeightModeHigher:
				switch currentHeightMode {
				case HeightModeHigher:
					// higher -> higher
					if lastOnePosY-lastOneOfLastOnePosY >= 2 {
						blockMatrix[x][z].PosY = lastOnePosY + 1
					} else {
						blockMatrix[x][z].PosY = lastOnePosY + 2
					}
				case HeightModeMiddle:
					// higher -> middle
					blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY
				case HeightModeLower:
					// higher -> lower
					if (flagX && flagZ) || (!flagX && !flagZ) {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY - 2
					} else {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY - 1
					}
				}
			case HeightModeMiddle:
				switch currentHeightMode {
				case HeightModeHigher:
					// middle -> higher
					if (flagX && flagZ) || (!flagX && !flagZ) {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY + 1
					} else {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY + 2
					}
				case HeightModeMiddle:
					// middle -> middle
					blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY
				case HeightModeLower:
					// middle -> lower
					if (flagX && flagZ) || (!flagX && !flagZ) {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY - 2
					} else {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY - 1
					}
				}
			case HeightModeLower:
				switch currentHeightMode {
				case HeightModeHigher:
					// lower -> higher
					if (flagX && flagZ) || (!flagX && !flagZ) {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY + 1
					} else {
						blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY + 2
					}
				case HeightModeMiddle:
					// lower -> middle
					blockMatrix[x][z].PosY = blockMatrix[x][z-1].PosY
				case HeightModeLower:
					// lower -> lower
					if lastOneOfLastOnePosY-lastOnePosY == 1 {
						blockMatrix[x][z].PosY = lastOnePosY - 2
					} else {
						blockMatrix[x][z].PosY = lastOnePosY - 1
					}
				}
			}
		}
	}

	minPosY := int16(384)
	for x := range 128 {
		for z := range 129 {
			minPosY = min(minPosY, blockMatrix[x][z].PosY)
		}
	}
	for x := range 128 {
		for z := range 129 {
			blockMatrix[x][z].PosY -= minPosY
			blockMatrix[x][z].PosY += int16(basePos[1])
		}
	}

	return [3]int32{
		basePos[0],
		basePos[1],
		basePos[2] - 1,
	}, blockMatrix
}
