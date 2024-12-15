package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient   *mongo.Client
	elasticClient *elastic.Client
)

func main() {
	ctx := context.Background()
	// Initialize MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	// mongoURI := "mongodb://localhost:27017"
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Initialize Elasticsearch
	elasticURI := os.Getenv("ELASTIC_URI")
	// elasticURI := "http://localhost:9200"
	elasticClient, err = elastic.NewClient(elastic.SetURL(elasticURI))
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	// http.HandleFunc("/", handleRequest)

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Query MongoDB
	collection := mongoClient.Database("exampledb").Collection("example")
	_, err := collection.CountDocuments(context.TODO(), map[string]interface{}{})
	if err != nil {
		http.Error(w, "MongoDB query failed", http.StatusInternalServerError)
		log.Println("MongoDB query failed:", err)
		return
	}

	// Query Elasticsearch
	_, err = elasticClient.Search().Index("exampleindex").Do(context.TODO())
	if err != nil {
		http.Error(w, "Elasticsearch query failed", http.StatusInternalServerError)
		log.Println("Elasticsearch query failed:", err)
		return
	}

	// Respond
	duration := time.Since(start).Milliseconds()
	fmt.Fprintf(w, "Request handled in %d ms\n", duration)
}
