package queue

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisRdbClient *redis.Client

func SetupRedisRDB() {
	redisRdbClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func PublishRedisRDB(c *gin.Context) {
	ctx := context.Background()
	err := redisRdbClient.RPush(ctx, "test_queue", "Hello, redis_rdb").Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to publish"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message published to redis_rdb"})
}

func ConsumeRedisRDB() {
	ctx := context.Background()
	for {
		msg, err := redisRdbClient.BLPop(ctx, 0, "test_queue").Result()
		if err != nil {
			log.Fatalf("Error consuming from redis_rdb: %v", err)
		}
		log.Printf("redis_rdb received: %s", msg[1])
		updateRedisRDBCounter()
	}
}
