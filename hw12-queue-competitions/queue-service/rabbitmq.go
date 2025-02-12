package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection
var rabbitChannel *amqp.Channel
var rabbitQueue amqp.Queue

func setupRabbitMQ() {
	var err error
	rabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	rabbitQueue, err = rabbitChannel.QueueDeclare(
		"test_queue", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
}

func publishRabbitMQ(c *gin.Context) {
	body := "Hello, RabbitMQ!"
	err := rabbitChannel.Publish(
		"", rabbitQueue.Name, false, false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(body)},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to publish"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message published to RabbitMQ!"})
}

func consumeRabbitMQ() {
	msgs, err := rabbitChannel.Consume(rabbitQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		log.Printf("RabbitMQ received: %s", msg.Body)
	}
}
