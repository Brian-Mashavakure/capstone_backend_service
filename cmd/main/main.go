package main

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/scans-apis/scans-routes"
	"github.com/gin-gonic/gin"
)

func main() {
	database.DatabaseConnect()
	router := gin.Default()

	scans_routes.ScanRoutes(router)

	router.Run("localhost:8080")

}
