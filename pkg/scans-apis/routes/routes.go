package routes

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/scans-apis/handlers"
	"github.com/gin-gonic/gin"
)

func ScanRoutes(router *gin.Engine) {
	api := router.Group("capstone/api")

	api.POST("/postscan", handlers.PostScanHandler)

	api.GET("/getscans", handlers.GetScansHandler)

}
