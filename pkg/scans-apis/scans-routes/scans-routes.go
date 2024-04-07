package scans_routes

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/scans-apis/scans-handlers"
	"github.com/gin-gonic/gin"
)

func ScanRoutes(router *gin.Engine) {
	api := router.Group("capstone/api")

	api.POST("/postscan", scans_handlers.PostScanHandler)

	api.GET("/getscans", scans_handlers.GetScansHandler)

}
