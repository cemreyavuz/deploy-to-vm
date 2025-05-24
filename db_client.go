package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type DBClient struct {
	Client *mongo.Client
}

type DBClientInterface interface {
	Connect()
}

func (dbClient *DBClient) Connect() {
	username := os.Getenv("MONGO_DB_USERNAME")
	password := os.Getenv("MONDO_DB_PASSWORD")
	cluster := os.Getenv("MONGO_DB_CLUSTER")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	connectionURI := "mongodb+srv://" + username + ":" + password + "@" + cluster + "/?retryWrites=true&w=majority"
	opts := options.Client().ApplyURI(connectionURI).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	// Set the client to the DBClient struct
	dbClient.Client = client

	//
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	log.Printf("Pinged your deployment. You successfully connected to MongoDB!")
}
