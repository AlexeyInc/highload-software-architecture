#!/bin/bash

# Queue selection
QUEUE_TYPE=$1  # "rabbitmq", "redis_rdb", or "redis_aof"

echo "Waiting for messages to start being published..."
while [ ! -f ./messages_started.lock ]; do
    sleep 0.2  # Wait for 200 milliseconds
done

echo "Consuming messages from $QUEUE_TYPE..."

while [ -f ./messages_started.lock ]; do
    case $QUEUE_TYPE in
        "rabbitmq")
            MESSAGE=$(curl -s -u "guest:guest" -X POST \
                "http://localhost:15672/api/queues/%2F/test_queue/get" \
                -H "Content-Type: application/json" \
                -d '{"count":1, "ackmode":"ack_requeue_false", "encoding":"auto", "truncate":500}')
            PAYLOAD=$(echo "$MESSAGE" | jq -r '.[0].payload // empty')
            if [ -n "$PAYLOAD" ]; then
                echo "Received from RabbitMQ: $PAYLOAD"
            fi
            ;;
        "redis_rdb")
            MESSAGE=$(docker exec redis_rdb redis-cli LPOP test_queue)
            if [ "$MESSAGE" != "(nil)" ] && [ -n "$MESSAGE" ]; then
                echo "Received from Redis RDB: $MESSAGE"
            fi
            ;;
        "redis_aof")
            MESSAGE=$(docker exec redis_aof redis-cli LPOP test_queue)
            if [ "$MESSAGE" != "(nil)" ] && [ -n "$MESSAGE" ]; then
                echo "Received from Redis AOF: $MESSAGE"
            fi
            ;;
        *)
            echo "Invalid queue type. Use 'rabbitmq', 'redis_rdb', or 'redis_aof'."
            exit 1
            ;;
    esac
done

echo "Finished consuming messages from $QUEUE_TYPE!"