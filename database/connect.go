package database

import (
    "context"
    "log"
    "os"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// Initialize the MongoDB client
func Connect() {
    mongoURI := os.Getenv("MONGO_DB_URI")
    if mongoURI == "" {
        log.Fatal("MONGO_DB_URI environment variable is not set")
    }

    clientOptions := options.Client().ApplyURI(mongoURI)
    var err error
    Client, err = mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Verify the connection
    err = Client.Ping(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Connected to MongoDB")
}
