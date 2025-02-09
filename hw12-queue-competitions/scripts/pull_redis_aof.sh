#!/bin/bash

# Redis AOF settings
REDIS_AOF_QUEUE="test_queue"

echo "Consuming message from Redis AOF..."
REDIS_AOF_MESSAGE=$(docker exec redis_aof redis-cli LPOP $REDIS_AOF_QUEUE)

if [ "$REDIS_AOF_MESSAGE" != "(nil)" ] && [ -n "$REDIS_AOF_MESSAGE" ]; then
    echo "Received from Redis AOF: $REDIS_AOF_MESSAGE"
else
    echo "No messages in Redis AOF queue."
fi