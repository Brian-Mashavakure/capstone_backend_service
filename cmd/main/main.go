package main

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/images-apis/images-routes"
	"github.com/gin-gonic/gin"
)

func main() {
	database.DatabaseConnect()
	router := gin.Default()

	images_routes.ImageRoutes(router)

	router.Run("localhost:8080")

}
