package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.ClusterClient
var ctx = context.Background()

func main() {
	// Initialize Redis Cluster Client
	redisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:6379",
			"localhost:6380",
			"localhost:6381",
			"localhost:6382",
			"localhost:6383",
			"localhost:6384"},
	})

	preloadKeys()

	// Eviction Testing Endpoints
	http.HandleFunc("/evict/lru", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-lru") })
	http.HandleFunc("/evict/random", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-random") })
	http.HandleFunc("/evict/ttl", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "volatile-ttl") })

	// Probabilistic Cache Clearing Endpoints
	http.HandleFunc("/cache/recompute", handleRecomputeCache)
	http.HandleFunc("/cache/probabilistic", handleProbabilisticCache)

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func preloadKeys() {
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := rand.Intn(10000)
		ttl := time.Duration(rand.Intn(300)) * time.Second // Random TTL up to 5 minutes
		if err := redisClient.Set(ctx, key, value, ttl).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			return
		}
	}
	log.Println("Preloaded 10,000 keys into Redis Cluster.")
}

func testEviction(w http.ResponseWriter, policy string) {
	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("test-key:%d", i)
		if err := redisClient.Set(ctx, key, i, 0).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			return
		}
	}

	w.Write([]byte(fmt.Sprintf("✅ Tested eviction strategy: %s", policy)))
}

func handleRecomputeCache(w http.ResponseWriter, r *http.Request) {
	key := "user:123"

	// Simulated expensive computation function
	recomputeFunc := func() string {
		time.Sleep(2 * time.Second) // Simulate slow DB call
		return fmt.Sprintf("UserData-%d", rand.Intn(10000))
	}

	val := getCachedValueWithRecompute(key, recomputeFunc)
	w.Write([]byte(fmt.Sprintf("✅ Cache Value (Recompute): %s", val)))
}

func handleProbabilisticCache(w http.ResponseWriter, r *http.Request) {
	key := "user:123"

	// Simulated expensive computation function
	recomputeFunc := func() string {
		time.Sleep(2 * time.Second) // Simulate slow DB call
		return fmt.Sprintf("UserData-%d", rand.Intn(10000))
	}

	val := getCachedValueWithProbabilisticTTL(key, recomputeFunc)
	w.Write([]byte(fmt.Sprintf("✅ Cache Value (Probabilistic): %s", val)))
}

func getCachedValueWithRecompute(key string, recomputeFunc func() string) string {
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		newValue := recomputeFunc()
		redisClient.Set(ctx, key, newValue, time.Minute)
		return newValue
	}

	go func() {
		newValue := recomputeFunc()
		redisClient.Set(ctx, key, newValue, time.Minute)
	}()

	return val
}

func getCachedValueWithProbabilisticTTL(key string, recomputeFunc func() string) string {
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		newValue := recomputeFunc()
		ttl := time.Minute + time.Duration(rand.Intn(30))*time.Second
		redisClient.Set(ctx, key, newValue, ttl)
		return newValue
	}
	return val
}
