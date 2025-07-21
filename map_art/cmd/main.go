package main

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/OmineDev/flowers-for-machines/client"
	"github.com/OmineDev/flowers-for-machines/game_control/game_interface"
	"github.com/OmineDev/flowers-for-machines/game_control/resources_control"
	"github.com/OmineDev/flowers-for-machines/map_art"

	_ "image/png"
)

var (
	c         *client.Client
	resources *resources_control.Resources
	api       *game_interface.GameInterface
)

func main() {
	SystemTestingLogin()
	defer func() {
		c.Conn().Close()
		time.Sleep(time.Second)
	}()
	doImport()
}

func doImport() {
	pngFileBytes, err := os.ReadFile("test.png")
	if err != nil {
		panic(err)
	}

	img, format, err := image.Decode(bytes.NewBuffer(pngFileBytes))
	if err != nil {
		panic(err)
	}
	if format != "png" {
		panic("doImport: Given image is not a png")
	}

	realImage := img.(*image.NRGBA)
	pixels := [128][128]color.RGBA{}
	if len(realImage.Pix) != 128*128*4 {
		panic("doImport: Given image must be 128*128 pixels")
	}

	idx := 0
	for z := range 128 {
		for x := range 128 {
			pixels[x][z] = color.RGBA{
				realImage.Pix[idx],
				realImage.Pix[idx+1],
				realImage.Pix[idx+2],
				realImage.Pix[idx+3],
			}
			idx += 4
		}
	}

	counter := 0
	importPos, result := map_art.GenerateMapArtStructure(
		[3]int32{-4416, -30, 6976},
		pixels,
	)
	for relativeX := range len(result) {
		for relativeZ := range len(result[relativeX]) {
			block := result[relativeX][relativeZ]

			realX := importPos[0] + int32(relativeX)
			realZ := importPos[2] + int32(relativeZ)

			err = api.SetBlock().SetBlockAsync(
				[3]int32{realX, int32(block.PosY), realZ},
				block.BlockMode.BlockName,
				"[]",
			)
			if err != nil {
				panic(err)
			}

			counter++
			if counter >= 20 {
				time.Sleep(time.Second / 20)
				counter = 0
			}
		}
	}
}
