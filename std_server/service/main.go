package main

import (
	"flag"
	"log"

	service "github.com/OmineDev/flowers-for-machines/std_server/service/src"
)

var (
	rentalServerCode     *string
	rentalServerPasscode *string
	authServerAddress    *string
	authServerToken      *string
	standardServerPort   *int
	consoleDimensionID   *int
	consoleCenterX       *int
	consoleCenterY       *int
	consoleCenterZ       *int
)

func init() {
	rentalServerCode = flag.String("rsn", "", "The rental server number.")
	rentalServerPasscode = flag.String("rsp", "", "The pass code of the rental server.")
	authServerAddress = flag.String("asa", "", "The auth server address.")
	authServerToken = flag.String("ast", "", "The auth server token.")
	standardServerPort = flag.Int("ssp", 0, "The server port to running.")
	consoleDimensionID = flag.Int("cdi", 0, "The dimension ID of the console. (e.g. overworld = 0, nether = 1, end = 2, dmT = T, etc.)")
	consoleCenterX = flag.Int("ccx", 0, "The X position of the center of the console.")
	consoleCenterY = flag.Int("ccy", 0, "The Y position of the center of the console.")
	consoleCenterZ = flag.Int("ccz", 0, "The Z position of the center of the console.")

	flag.Parse()
	if len(*rentalServerCode) == 0 {
		log.Fatalln("Please provide your rental server number.\n\te.g. -rsn=\"123456\"")
	}
	if len(*authServerAddress) == 0 {
		log.Fatalln("Please provide your auth server address.\n\te.g. -asa=\"http://127.0.0.1\"")
	}
	if *standardServerPort == 0 {
		log.Fatalln("Please provide the server port to running.\n\te.g. -ssp=0")
	}
}

func main() {
	service.RunServer(
		*rentalServerCode,
		*rentalServerPasscode,
		*authServerAddress,
		*authServerToken,
		*standardServerPort,
		*consoleDimensionID,
		*consoleCenterX,
		*consoleCenterY,
		*consoleCenterZ,
	)
}
