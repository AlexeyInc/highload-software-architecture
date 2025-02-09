#!/bin/bash

# Redis RDB settings
REDIS_RDB_QUEUE="test_queue"

# Define message payload
MESSAGE="Hello, Redis RDB!"

echo "Publishing message to Redis RDB..."
docker exec redis_rdb redis-cli RPUSH $REDIS_RDB_QUEUE "$MESSAGE"

echo "Message sent to Redis RDB!"