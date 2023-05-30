package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/rodblg/Challenge-API-Go/controllers"
)

func TransactionRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.GET("/transactions", controller.GetTransactions())
	incomingRoutes.GET("/transactions/:transaction_id", controller.GetTransaction())
	incomingRoutes.POST("/transactions", controller.CreateTransaction())

	//incomingRoutes.PATCH("/transactions/:transaction_id", controller.UpdateTransaction())
}
