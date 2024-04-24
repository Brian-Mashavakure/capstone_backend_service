package images_routes

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/images-apis/images-handlers"
	"github.com/gin-gonic/gin"
)

func ImageRoutes(router *gin.Engine) {
	api := router.Group("capstone/api")

	api.POST("/postimage", images_handlers.PostImageHandler)

	api.GET("/getimages/:username", images_handlers.GetImagesHandler)

}
