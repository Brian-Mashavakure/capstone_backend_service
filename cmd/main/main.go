package main

import (
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/database"
	"github.com/Brian-Mashavakure/capstone_backend_service/pkg/scans-apis/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	database.DatabaseConnect()
	router := gin.Default()

	routes.ScanRoutes(router)

	router.Run("192.168.10.161:8080")

}
