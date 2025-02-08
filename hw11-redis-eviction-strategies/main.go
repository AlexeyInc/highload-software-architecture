package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
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
	locks       = sync.Map{}
	redisClient *redis.ClusterClient
	ctx         = context.Background()

	defultKeysTTL = time.Minute * 10
	customTTLSec  int
	beta          = 1.0
)

func main() {
	redisNodes := getRedisClusterNodes()
	if len(redisNodes) == 0 {
		log.Fatal("No Redis nodes found in environment variable REDIS_CLUSTER_NODES")
	}

	redisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: redisNodes,
	})

	r := mux.NewRouter()

	r.HandleFunc("/preloadKeys", func(w http.ResponseWriter, r *http.Request) { preloadKeys() })
	r.HandleFunc("/evict/lru", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-lru") })
	r.HandleFunc("/evict/lfu", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-lfu") })
	r.HandleFunc("/evict/random", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "allkeys-random") })
	r.HandleFunc("/evict/ttl", func(w http.ResponseWriter, r *http.Request) { testEviction(w, "volatile-ttl") })

	r.HandleFunc("/fetch/db/{key}", fetchFromDB).Methods("GET")
	r.HandleFunc("/fetch/cache/{key}", fetchWithCache).Methods("GET")
	r.HandleFunc("/fetch/external/{key}", fetchWithExternalRecomputation).Methods("GET")
	r.HandleFunc("/fetch/probabilistic/{key}", fetchWithProbabilisticExpiration).Methods("GET")
	r.HandleFunc("/set/{key}/{value}/{ttl}", setCacheValue).Methods("POST")
	r.HandleFunc("/delete/{key}", deleteCacheKey).Methods("DELETE")

	go monitorEvictions()

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func monitorEvictions() {
	err := redisClient.ConfigSet(ctx, "notify-keyspace-events", "KEA").Err()
	if err != nil {
		log.Fatalf("Failed to enable keyspace notifications: %v", err)
	}

	pubsub := redisClient.PSubscribe(ctx, "__keyevent@0__:evicted")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		log.Printf("Evicted key: %s", msg.Payload)
	}
}

func getRedisClusterNodes() []string {
	nodesEnv := os.Getenv("REDIS_CLUSTER_NODES")
	if nodesEnv == "" {
		return nil
	}
	return strings.Split(nodesEnv, ",")
}

func preloadKeys() {
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := fmt.Sprintf("V%d", i)
		if err := redisClient.Set(ctx, key, value, defultKeysTTL).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			return
		}
		log.Printf("Set key %s: %v", key, value)
	}
	log.Println("Preloaded 10000 keys into Redis Cluster.")
}

func testEviction(w http.ResponseWriter, policy string) {
	err := redisClient.Do(ctx, "CONFIG", "SET", "maxmemory-policy", policy).Err()
	if err != nil {
		log.Printf("Failed to set Do eviction policy %s: %v", policy, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to set eviction policy: %v", err)))
		return
	}

	log.Printf("Set eviction policy to: %s", policy)

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("test-key:%d-%s", i, policy)
		time.Sleep(time.Millisecond)
		if err := redisClient.Set(ctx, key, i, defultKeysTTL).Err(); err != nil {
			log.Printf("Failed to set key %s: %v", key, err)
			time.Sleep(time.Second)
			continue
		}
	}

	w.Write([]byte(fmt.Sprintf("Tested eviction strategy: %s", policy)))
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
	w.Write([]byte(fmt.Sprintf("fetchFromDB: key:%s -> value: %s", key, value)))
}

// Method 2: Fetch from cache, fallback to DB on miss
func fetchWithCache(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		val, err = mockDBQuery(key)
		log.Printf("WithCache made DBQuery %s: %v", key, val)

		if err == nil {
			redisClient.Set(ctx, key, val, time.Duration(customTTLSec)*time.Second)
		}
	} else if err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("fetchWithCache: key:%s -> value: %s", key, val)))
}

// Method 3: External recomputation with locking
func fetchWithExternalRecomputation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if val, err := redisClient.Get(ctx, key).Result(); err == nil {
		w.Write([]byte(val))
		return
	}

	mutex, _ := locks.LoadOrStore(key, &sync.Mutex{})
	lock := mutex.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	// Double check after acquiring lock
	if val, err := redisClient.Get(ctx, key).Result(); err == nil {
		w.Write([]byte(val))
		return
	}

	val, err := mockDBQuery(key)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	log.Printf("ExternalRecomputation made DBQuery %s: %v", key, val)

	redisClient.Set(ctx, key, val, time.Duration(customTTLSec)*time.Second)
	w.Write([]byte(fmt.Sprintf("fetchWithExternalRecomputation: key:%s -> value: %s", key, val)))
}

// Method 4: Probabilistic early expiration
func fetchWithProbabilisticExpiration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	val, ttl, err := getCacheValueWithTTL(key)
	if err == redis.Nil || shouldRecompute(ttl, customTTLSec) {
		mutex, _ := locks.LoadOrStore(key, &sync.Mutex{})
		lock := mutex.(*sync.Mutex)
		lock.Lock()
		defer lock.Unlock()

		// Double check after acquiring lock
		if val, err := redisClient.Get(ctx, key).Result(); err == nil {
			w.Write([]byte(val))
			return
		}

		val, err = mockDBQuery(key)
		if err == nil {
			redisClient.Set(ctx, key, val, time.Duration(customTTLSec)*time.Second)
		}
		log.Printf("ProbabilisticExpiration made DBQuery %s: %v", key, val)
	} else if err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("fetchWithProbabilisticExpiration: key:%s -> value: %s", key, val)))
}

func shouldRecompute(ttl time.Duration, customTTLSec int) bool {
	if ttl <= time.Millisecond {
		return true
	}

	remainingRatio := float64(ttl.Seconds()) / float64(customTTLSec) // 0.0 - 1.0
	probability := 1 - math.Exp(-beta*(1-remainingRatio))            // Probabilistic early expiration formula

	return rand.Float64() < probability
}

func getCacheValueWithTTL(key string) (string, time.Duration, error) {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}

	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		return val, 0, err
	}
	return val, ttl, nil
}

func setCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]
	customTTLSec, _ = strconv.Atoi(vars["ttl"])
	redisClient.Set(ctx, key, value, time.Duration(customTTLSec)*time.Second)
	w.Write([]byte(fmt.Sprintf("setCacheValue: key:%s -> value: %s. Ok", key, value)))
}

func deleteCacheKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := redisClient.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed delete key %s: %v", key, err)
		http.Error(w, fmt.Sprintf("Failed delete key: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Deleted key: %s", key)))
}
