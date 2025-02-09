#!/bin/bash

# Redis RDB settings
REDIS_RDB_QUEUE="test_queue"

echo "Consuming message from Redis RDB..."
REDIS_RDB_MESSAGE=$(docker exec redis_rdb redis-cli LPOP $REDIS_RDB_QUEUE)

if [ "$REDIS_RDB_MESSAGE" != "(nil)" ] && [ -n "$REDIS_RDB_MESSAGE" ]; then
    echo "Received from Redis RDB: $REDIS_RDB_MESSAGE"
else
    echo "No messages in Redis RDB queue."
fi