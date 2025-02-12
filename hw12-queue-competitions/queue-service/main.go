package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup RabbitMQ, Redis RDB, Redis AOF
	setupRabbitMQ()
	setupRedisRDB()
	setupRedisAOF()

	// Run Consumers in Goroutines
	go consumeRabbitMQ()
	go consumeRedisRDB()
	go consumeRedisAOF()

	// HTTP Server for Publishing
	r := gin.Default()
	r.GET("/publish/rabbitmq", publishRabbitMQ)
	r.GET("/publish/redis_rdb", publishRedisRDB)
	r.GET("/publish/redis_aof", publishRedisAOF)

	log.Println("Server started on :8080")
	r.Run(":8080")
}
