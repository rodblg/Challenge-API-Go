package controller

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rodblg/Challenge-API-Go/database"
	"github.com/rodblg/Challenge-API-Go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var transactionCollection *mongo.Collection = database.OpenCollection(database.Client, "transaction")
var validate = validator.New()

func GetTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var transactionId = c.Param("transaction_id")
		var transaction models.Transaction

		err := transactionCollection.FindOne(ctx, bson.M{"_id": transactionId}).Decode(&transaction)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occured while listing items"})
			return
		}
		c.JSON(http.StatusOK, transaction)
	}
}

func GetTransactions() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		userId := c.GetString("uid")
		//usertId, _ := primitive.ObjectIDFromHex(userId)
		matchStage := bson.D{{"$match", bson.D{{"user_id", userId}}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "_id"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}

		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_count", 1},
					{"data", 1},
				}}}

		result, err := transactionCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing transactions items"})
			return
		}

		var allTransactions []bson.M
		if err = result.All(ctx, &allTransactions); err != nil {
			fmt.Println(err)
			log.Fatal(err)

		}
		val := allTransactions[0]
		valsg := val["total_count"]
		fmt.Println(valsg)

		c.JSON(http.StatusOK, allTransactions)

	}

}

func CreateTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var transaction models.Transaction
		var user models.User
		user_id := c.GetString("uid")

		if err := c.BindJSON(&transaction); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		transaction.User_id = user_id
		validationErr := validate.Struct(transaction)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		}
		
		err := userCollection.FindOne(ctx, bson.M{"user_id": user_id}).Decode(&user)
		defer cancel()
		if err != nil {
			msg := "user was not found"
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
		}

		transaction.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		transaction.ID = primitive.NewObjectID()
		var num = toFixed(*transaction.Value, 2)
		transaction.Value = &num

		result, insertErr := transactionCollection.InsertOne(ctx, transaction)
		defer cancel()
		if insertErr != nil {
			msg := fmt.Sprintf("transaction item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		filter := bson.D{primitive.E{Key:"user_id", Value: user_id}}
		update := bson.D{{"$push", bson.D{primitive.E{Key:"movements", Value: transaction}}}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Println(err)
			msg := "could not update transaction for user"
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
		}
		
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
