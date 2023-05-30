package main

import (
	"os"

	"github.com/gin-gonic/gin"
	middleware "github.com/rodblg/Challenge-API-Go/middleware"
	routes "github.com/rodblg/Challenge-API-Go/routes"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)

	router.Use(middleware.Authentication())
	routes.TransactionRoutes(router)

	router.Run(":" + port)
}
