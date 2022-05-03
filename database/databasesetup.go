package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var (
	Client = DBSet()
)

func DBSet() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("failed to connect to mongodb")
		return nil
	}

	log.Println("successfully connected to mongodb")
	return client
}

func UserData(client *mongo.Client, collectionName string) (userCollection *mongo.Collection) {
	userCollection = client.Database("Ecommerce").Collection(collectionName)
	return userCollection
}

func ProductData(client *mongo.Client, collectionName string) (productCollection *mongo.Collection) {
	productCollection = client.Database("Ecommerce").Collection(collectionName)
	return productCollection
}
