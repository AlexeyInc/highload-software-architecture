#!/bin/bash

# Redis AOF settings
REDIS_AOF_QUEUE="test_queue"

# Define message payload
MESSAGE="Hello, Redis AOF!"

echo "Publishing message to Redis AOF..."
docker exec redis_aof redis-cli RPUSH $REDIS_AOF_QUEUE "$MESSAGE"

echo "Message sent to Redis AOF!"