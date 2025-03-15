package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient     *mongo.Client
	mongoDB         string
	mongoCollection string
)

func main() {
	var (
		ctx = context.Background()
		err error
	)
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB = os.Getenv("MONGO_DB")
	mongoCollection = os.Getenv("MONGO_DB_COLLECTION")

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	err = createDBAndCollection(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database and collection: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", RecoveryMiddleware(mux)))
}

func createDBAndCollection(ctx context.Context) error {
	err := mongoClient.Database(mongoDB).CreateCollection(ctx, mongoCollection)
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}
	log.Println("Database and collection initialized")
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	collection := mongoClient.Database(mongoDB).Collection(mongoCollection)

	uniqueKey := time.Now().UnixNano() % 100 // Simulate 100 keys to increase contention

	filter := bson.M{"key": uniqueKey}
	update := bson.M{
		"$set": bson.M{
			"timestamp":  time.Now(),
			"ping":       "pong",
			"updated":    true,
			"updated_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true) // ensure an upsert occurs
	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		http.Error(w, "collection.UpdateOne failed", http.StatusInternalServerError)
		return
	}

	time.Sleep(10 * time.Millisecond) // artificial delay to simulate processing time

	fmt.Fprintln(w, "MongoDB Insert or Update Successful")

	duration := time.Since(start).Milliseconds()
	fmt.Fprintf(w, "Request handled in %d ms\n", duration)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
