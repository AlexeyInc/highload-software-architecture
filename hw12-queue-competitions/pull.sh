#!/bin/bash

# RabbitMQ Push
RABBITMQ_USER="user"
RABBITMQ_PASS="password"
RABBITMQ_QUEUE="test_queue"
RABBITMQ_EXCHANGE=""
MESSAGE="Hello from Bash"

echo "Consuming message from RabbitMQ..."
curl -u $RABBITMQ_USER:$RABBITMQ_PASS -X GET http://localhost:15672/api/queues/%2F/$RABBITMQ_QUEUE/get -H "Content-Type: application/json" -d '{"count":1, "ackmode":"ack_requeue_false", "encoding":"auto"}'

echo "Consuming message from Redis RDB..."
redis-cli -h localhost -p 6379 XREAD COUNT 1 STREAMS stream_rdb 0

echo "Consuming message from Redis AOF..."
redis-cli -h localhost -p 6380 XREAD COUNT 1 STREAMS stream_aof 0

echo "Message consumption completed!"
