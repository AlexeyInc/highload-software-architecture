#!/bin/bash

REDIS_RDB_QUEUE="test_queue"

MESSAGE="Hello, Redis RDB!"

echo "Publishing message to Redis RDB..."
docker exec redis_rdb redis-cli RPUSH $REDIS_RDB_QUEUE "$MESSAGE"

echo "Message sent to Redis RDB!"