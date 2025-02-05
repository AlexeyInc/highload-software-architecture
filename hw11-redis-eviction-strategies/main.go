package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

var (
	rdb          *redis.ClusterClient
	locks        = sync.Map{}
	expireJitter = 10 // Jitter in seconds for early expiration
	redisClient  *redis.ClusterClient
	ctx          = context.Background()
)

func main() {
	redisNodes := getRedisClusterNodes()
	if len(redisNodes) == 0 {
		log.Fatal("No Redis nodes found in environment variable REDIS_CLUSTER_NODES")
	}

	redisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: redisNodes,
	})

	preloadKeys()

	r := mux.NewRouter()

	r.HandleFunc("/evict/lru", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-lru") })
	r.HandleFunc("/evict/lfu", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-lfu") })
	r.HandleFunc("/evict/random", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-random") })
	r.HandleFunc("/evict/ttl", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "volatile-ttl") })

	r.HandleFunc("/fetch/db/{key}", fetchFromDB).Methods("GET")
	r.HandleFunc("/fetch/cache/{key}", fetchWithCache).Methods("GET")
	r.HandleFunc("/fetch/external/{key}", fetchWithExternalRecomputation).Methods("GET")
	r.HandleFunc("/fetch/probabilistic/{key}", fetchWithProbabilisticExpiration).Methods("GET")
	r.HandleFunc("/set/{key}/{value}/{ttl}", setCacheValue).Methods("POST")

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getRedisClusterNodes() []string {
	nodesEnv := os.Getenv("REDIS_CLUSTER_NODES")
	if nodesEnv == "" {
		return nil
	}
	return strings.Split(nodesEnv, ",")
}

func preloadKeys() {
	for i := 0; i < 25000; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := rand.Intn(1000)
		ttl := (time.Duration(rand.Intn(3)) + 1) * time.Minute // Random 1 min < TTL <= 4 min
		if err := redisClient.Set(ctx, key, value, ttl).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			return
		}
	}
	log.Println("Preloaded 25000 keys into Redis Cluster.")
}

func testEviction(w http.ResponseWriter, policy string) {
	err := redisClient.ConfigSet(ctx, "maxmemory-policy", policy).Err()
	if err != nil {
		log.Printf("Failed to set ConfigSet eviction policy %s: %v", policy, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to set eviction policy: %v", err)))
		return
	}

	err = redisClient.Do(ctx, "CONFIG", "SET", "maxmemory-policy", policy).Err()
	if err != nil {
		log.Printf("Failed to set Do eviction policy %s: %v", policy, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to set eviction policy: %v", err)))
		return
	}

	log.Printf("Set eviction policy to: %s", policy)

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("test-key:%d", i)
		time.Sleep(time.Millisecond * 5)
		if err := redisClient.Set(ctx, key, i, 0).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			return
		}
	}

	w.Write([]byte(fmt.Sprintf("Tested eviction strategy: %s", policy)))
}

func init() {
	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:6379"},
	})
}

func mockDBQuery(key string) (string, error) {
	time.Sleep(500 * time.Millisecond) // Simulating DB latency
	return "value_from_db_" + key, nil
}

// Method 1: Directly fetching from DB
func fetchFromDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value, err := mockDBQuery(key)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(value))
}

// Method 2: Fetch from cache, fallback to DB on miss
func fetchWithCache(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		val, err = mockDBQuery(key)
		if err == nil {
			rdb.Set(ctx, key, val, 10*time.Second)
		}
	} else if err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(val))
}

// Method 3: External recomputation with locking
func fetchWithExternalRecomputation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if val, err := rdb.Get(ctx, key).Result(); err == nil {
		w.Write([]byte(val))
		return
	}

	mutex, _ := locks.LoadOrStore(key, &sync.Mutex{})
	lock := mutex.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	// Double check after acquiring lock
	if val, err := rdb.Get(ctx, key).Result(); err == nil {
		w.Write([]byte(val))
		return
	}

	val, err := mockDBQuery(key)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	rdb.Set(ctx, key, val, 10*time.Second)
	w.Write([]byte(val))
}

// Method 4: Probabilistic early expiration
func fetchWithProbabilisticExpiration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || shouldRecompute(key) {
		val, err = mockDBQuery(key)
		if err == nil {
			rdb.Set(ctx, key, val, time.Duration(10+rand.Intn(expireJitter))*time.Second)
		}
	} else if err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(val))
}

func shouldRecompute(key string) bool {
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		return true
	}
	threshold := float64(ttl.Seconds()) / 10.0 // Expire when 10% time left
	return rand.Float64() < threshold
}

// Endpoint to set values in cache
func setCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]
	ttlSec, _ := strconv.Atoi(vars["ttl"])
	rdb.Set(ctx, key, value, time.Duration(ttlSec)*time.Second)
	w.Write([]byte("OK"))
}
