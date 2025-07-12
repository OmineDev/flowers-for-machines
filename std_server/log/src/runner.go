package log

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func initRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", Root)
	router.POST("/set_auth_key", SetAuthKey)
	router.POST("/log_record", LogRecord)
	router.POST("/log_review", LogReview)
	router.POST("/log_finish_review", LogFinishReview)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return router
}

func RunServer() {
	router := initRouter()
	router.Run(":8080")
}
