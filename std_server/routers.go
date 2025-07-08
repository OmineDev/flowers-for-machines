package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/check_alive", CheckAlive)
	router.GET("/process_exit", ProcessExist)
	router.GET("/change_console_position", ChangeConsolePosition)

	router.POST("/place_nbt_block", PlaceNBTBlock)
	router.POST("/place_large_chest", PlaceLargeChest)
	router.POST("/place_water_logged_block", PlaceWaterloggedBlock)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}

func RunServer() {
	router := InitRouter()
	router.Run(fmt.Sprintf(":%d", *standardServerPort))
}
