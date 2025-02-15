package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"queue-service/queue"
)

func main() {
	queue.SetupRabbitMQ()
	queue.SetupRedisRDB()
	queue.SetupRedisAOF()

	go queue.ConsumeRabbitMQ()
	go queue.ConsumeRedisRDB()
	go queue.ConsumeRedisAOF()

	r := gin.Default()
	r.GET("/publish/rabbitmq", queue.PublishRabbitMQ)
	r.GET("/publish/redis_rdb", queue.PublishRedisRDB)
	r.GET("/publish/redis_aof", queue.PublishRedisAOF)

	r.GET("/msg/counter/:queue_name", getConsumedMessagesCount)

	log.Println("Server started on :8080")
	r.Run(":8080")
}

func getConsumedMessagesCount(c *gin.Context) {
	queueName := c.Param("queue_name")
	reset := c.DefaultQuery("reset", "")

	consumedMessages := 0
	switch queueName {
	case "rabbitmq":
		if reset == "y" || reset == "yes" {
			queue.CounterRabbitMQ = 0
		}
		consumedMessages = queue.CounterRabbitMQ
	case "redis_rdb":
		if reset == "y" || reset == "yes" {
			queue.CounterRedisRDB = 0
		}
		consumedMessages = queue.CounterRedisRDB
	case "redis_aof":
		if reset == "y" || reset == "yes" {
			queue.CounterRedisAOF = 0
		}
		consumedMessages = queue.CounterRedisAOF
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
		return
	}

	log.Printf("Total Messages consumed in %s: %d", queueName, consumedMessages)

	c.JSON(http.StatusOK, fmt.Sprintf("Message consumed in %s: %d", queueName, consumedMessages))
}
