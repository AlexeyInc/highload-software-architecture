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
		Addr: "localhost:6380",
	})
}

func PublishRedisAOF(c *gin.Context) {
	ctx := context.Background()
	err := redisAofClient.RPush(ctx, "test_queue", "Hello, redis_aof").Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to publish"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message published to redis_aof"})
}

func ConsumeRedisAOF() {
	ctx := context.Background()
	for {
		msg, err := redisAofClient.BLPop(ctx, 0, "test_queue").Result()
		if err != nil {
			log.Fatalf("Error consuming from redis_aof: %v", err)
		}
		log.Printf("redis_aof received: %s", msg[1])
		updateRedisAOFCounter()
	}
}
