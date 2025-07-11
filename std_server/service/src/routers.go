package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func initRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/check_alive", CheckAlive)
	router.GET("/process_exit", ProcessExist)

	router.POST("/change_console_position", ChangeConsolePosition)
	router.POST("/place_nbt_block", PlaceNBTBlock)
	router.POST("/place_large_chest", PlaceLargeChest)
	router.POST("/get_nbt_block_hash", GetNBTBlockHash)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}

func runHttpServer(standardServerPort int) {
	router := initRouter()
	router.Run(fmt.Sprintf(":%d", standardServerPort))
}
