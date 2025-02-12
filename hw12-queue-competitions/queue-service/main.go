package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"queue-service/queue"
)

func main() {
	// Setup RabbitMQ, Redis RDB, Redis AOF
	queue.SetupRabbitMQ()
	queue.SetupRedisRDB()
	queue.SetupRedisAOF()

	// Run Consumers in Goroutines
	go queue.ConsumeRabbitMQ()
	go queue.ConsumeRedisRDB()
	go queue.ConsumeRedisAOF()

	// HTTP Server for Publishing
	r := gin.Default()
	r.GET("/publish/rabbitmq", queue.PublishRabbitMQ)
	r.GET("/publish/redis_rdb", queue.PublishRedisRDB)
	r.GET("/publish/redis_aof", queue.PublishRedisAOF)

	log.Println("Server started on :8080")
	r.Run(":8080")
}
