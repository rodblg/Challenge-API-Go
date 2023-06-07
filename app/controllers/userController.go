package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"html/template"
	"bytes"
	"sync"

	"github.com/gin-gonic/gin"
	database "github.com/rodblg/Challenge-API-Go/database"
	helper "github.com/rodblg/Challenge-API-Go/helpers"
	"github.com/rodblg/Challenge-API-Go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	//"github.com/jordan-wright/email"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

type Movement struct{
	Name    string
	Value	float64
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//validationErr := validate.Struct(user)
		//if validationErr != nil {
		//	c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		//	return
		//}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for user email"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for user phone"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		user.Movements = make([]models.Transaction, 0)

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(*user.Email)

		err := userCollection.FindOne(ctx, bson.M{"email": *user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found, incorrect credentials"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var userId = c.GetString("uid")
		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error finding user"})
		}

		if len(user.Movements) != 0 {

		matchStage := bson.D{{"$match", bson.D{{"user_id", userId}}}}
		unwindStage := bson.D{{"$unwind", bson.D{primitive.E{Key:"path", Value:"$movements"}}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", "movements"}, {"total_count_movements", bson.D{{"$sum", 1}}}, {"current_balance", bson.D{primitive.E{Key:"$sum", Value:"$movements.value"}}}, {"data", bson.D{primitive.E{Key:"$push", Value:"$movements"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
			{"_id", 0},
			{"total_count_movements", 1},
			{"current_balance", 1},
			{"data",1},
			}}}
		currentresult, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, unwindStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing transaction items"})
			return
		}
		var allTransactions []bson.M 
		if err = currentresult.All(ctx, &allTransactions); err != nil {

			fmt.Println(err)
			log.Fatal(err)
		}

		//fmt.Println(allTransactions)
	
		user.Balance = new(float64)
		fbalance := *user.Balance + allTransactions[0]["current_balance"].(float64)
		*user.Balance = fbalance
		filter2  := bson.D{{"user_id", userId}}
		update2 := bson.D{{"$set", bson.D{primitive.E{Key:"balance", Value: fbalance}}}}
		_, err = userCollection.UpdateOne(ctx, filter2, update2)
		defer cancel()
		if err != nil {
			log.Println(err)
		}
		
		allTransactions[0]["current_balance"] = fbalance
		//fmt.Printf("%T", allTransactions[0])
		c.JSON(http.StatusOK, allTransactions)
	}else{
		user.Balance = new(float64)

		filter  := bson.D{{"user_id", userId}}
		update := bson.D{{"$set", bson.D{primitive.E{"balance", *user.Balance}}}}
		_, err := userCollection.UpdateOne(ctx, filter, update)
		defer cancel()
		if err != nil {
			log.Println(err)
		}
		c.JSON(http.StatusOK, user)
	}

	}
}

func GetStatement() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		user_id := c.GetString("uid")
		var user models.User 
	
		err := userCollection.FindOne(ctx, bson.M{"user_id": user_id}).Decode(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"could not find user"})
		}

		//if len(user.Movements) != 0 {

		matchStage := bson.D{{"$match", bson.D{primitive.E{Key:"user_id", Value: user_id}}}}
		unwindStage := bson.D{{"$unwind", bson.D{primitive.E{Key:"path", Value:"$movements"}}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", "movements"}, {"total_count_movements", bson.D{{"$sum", 1}}}, {"current_balance", bson.D{primitive.E{Key:"$sum", Value:"$movements.value"}}}, {"data", bson.D{primitive.E{Key:"$push", Value:"$movements"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
			{"_id", 0},
			{"total_count_movements", 1},
			{"current_balance", 1},
			{"data",1},
			}}}
		currentresult, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, unwindStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing transaction items"})
			return
		}
		var allTransactions []bson.M 
		if err = currentresult.All(ctx, &allTransactions); err != nil {

			fmt.Println(err)
			log.Fatal(err)
		}
		user.Balance = new(float64)
		fbalance := *user.Balance + allTransactions[0]["current_balance"].(float64)
		allTransactions[0]["current_balance"] = fbalance
		
		var usertransactions []Movement
		for _, transaction := range allTransactions {
			dataArray := transaction["data"].(bson.A)
			for _, data := range dataArray {
				name := data.(bson.M)["name_movement"].(string)
				value := data.(bson.M)["value"].(float64)
				msg := new(Movement)
				msg.Name= name
				msg.Value= value
				

				usertransactions = append(usertransactions, *msg)
			
		}
		fmt.Print(usertransactions)

	}

		// Send the email concurrently
		var wg sync.WaitGroup
		wg.Add(1)	

		
		first_name := c.GetString("first_name")
		subject := fmt.Sprintf("Bank Statement for %s", first_name)
		body, err := getEmailBody("static/email-template.html", usertransactions)
		
		if err != nil {
			log.Println("Failed to read email template:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to read email template"})
			return
		}
		recipientEmail := c.GetString("email")

		go helper.SendEmail(subject, body, recipientEmail, &wg)


		c.JSON(http.StatusOK, gin.H{"message": "the statement has been sent"})	

			
	}
	
}


func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}

func getEmailBody(templateFile string, data []Movement)(string, error) {


    tmpl, err := template.ParseFiles(templateFile)
    if err != nil {
        return  "",err
    }
	var result bytes.Buffer
    err = tmpl.Execute(&result, data)
    if err != nil {
        return  "",err
    }

	return  result.String(), nil
    
}


