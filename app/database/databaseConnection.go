package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Some error occured while reading .env file")
	}
	
	MongoDB := os.Getenv("MONGO_URI")
	if MongoDB == "" {
		MONGO_INITDB_ROOT_USERNAME := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
		MONGO_INITDB_ROOT_PASSWORD := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
		MongoDB = fmt.Sprintf("mongodb://%s:%s@mongodb", MONGO_INITDB_ROOT_USERNAME, MONGO_INITDB_ROOT_PASSWORD)
	}


	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("connected to mongo")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("bank").Collection(collectionName)

	return collection

}
