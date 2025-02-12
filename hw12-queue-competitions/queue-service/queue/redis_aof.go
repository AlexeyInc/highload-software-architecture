package queue

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisAofClient *redis.Client

func SetupRedisAOF() {
	redisAofClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6380", // Change to match your Redis AOF instance
	})
}

func PublishRedisAOF(c *gin.Context) {
	ctx := context.Background()
	err := redisAofClient.RPush(ctx, "test_queue", "Hello, Redis AOF!").Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to publish"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message published to Redis AOF!"})
}

func ConsumeRedisAOF() {
	ctx := context.Background()
	for {
		msg, err := redisAofClient.BLPop(ctx, 0, "test_queue").Result()
		if err != nil {
			log.Fatalf("Error consuming from Redis AOF: %v", err)
		}
		log.Printf("Redis AOF received: %s", msg[1])
	}
}
