package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/rodblg/Challenge-API-Go/controllers"
	middleware "github.com/rodblg/Challenge-API-Go/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.POST("/users/signup", controller.SignUp())
	incomingRoutes.POST("/users/login", controller.Login())

	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.GET("/users/user", controller.GetUser())
	incomingRoutes.GET("/users/statement", controller.GetStatement())

}
