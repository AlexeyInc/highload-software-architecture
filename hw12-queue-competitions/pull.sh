#!/bin/bash

# RabbitMQ settings
RABBITMQ_HOST="localhost"
RABBITMQ_QUEUE="test_queue"
RABBITMQ_USER="guest"
RABBITMQ_PASS="guest"

# Redis settings
REDIS_RDB_QUEUE="test_queue"
REDIS_AOF_QUEUE="test_queue"

echo "Consuming message from RabbitMQ..."
MESSAGE=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" -X POST \
    "http://$RABBITMQ_HOST:15672/api/queues/%2F/$RABBITMQ_QUEUE/get" \
    -H "Content-Type: application/json" \
    -d '{"count":1, "ackmode":"ack_requeue_false", "encoding":"auto", "truncate":500}')

# Extract message payload
MESSAGE_PAYLOAD=$(echo "$MESSAGE" | jq -r '.[0].payload')

if [ "$MESSAGE_PAYLOAD" != "null" ]; then
    echo "Received from RabbitMQ: $MESSAGE_PAYLOAD"
else
    echo "No messages in RabbitMQ queue."
fi

echo "Consuming message from Redis RDB..."
REDIS_RDB_MESSAGE=$(docker exec redis_rdb redis-cli LPOP $REDIS_RDB_QUEUE)

if [ "$REDIS_RDB_MESSAGE" != "(nil)" ] && [ -n "$REDIS_RDB_MESSAGE" ]; then
    echo "Received from Redis RDB: $REDIS_RDB_MESSAGE"
else
    echo "No messages in Redis RDB queue."
fi

echo "Consuming message from Redis AOF..."
REDIS_AOF_MESSAGE=$(docker exec redis_aof redis-cli LPOP $REDIS_AOF_QUEUE)

if [ "$REDIS_AOF_MESSAGE" != "(nil)" ] && [ -n "$REDIS_AOF_MESSAGE" ]; then
    echo "Received from Redis AOF: $REDIS_AOF_MESSAGE"
else
    echo "No messages in Redis AOF queue."
fi

echo "Messages consumed from all queues!"