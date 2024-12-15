package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient   *mongo.Client
	elasticClient *elasticsearch.Client
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
	// elasticURI := os.Getenv("ELASTIC_URI")
	// elasticURI := "http://localhost:9200"
	elasticClient, err = elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	http.HandleFunc("/", handleRequest)

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	collection := mongoClient.Database("testdb").Collection("testcol")
	_, err := collection.InsertOne(nil, map[string]string{"ping": "pong", "timestamp": time.Now().String()})
	if err != nil {
		http.Error(w, "MongoDB Insert Failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "MongoDB Ping Successful")

	req := map[string]string{"ping": "pong", "timestamp": time.Now().String()}
	// Encode request body to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
	res, err := elasticClient.Index("testindex", bytes.NewReader(reqBody))
	if err != nil || res.IsError() {
		http.Error(w, "Elasticsearch Ping Failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "Elasticsearch Ping Successful")

	// Respond
	duration := time.Since(start).Milliseconds()
	fmt.Fprintf(w, "Request handled in %d ms\n", duration)
}
