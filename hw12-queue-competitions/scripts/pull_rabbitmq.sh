#!/bin/bash

RABBITMQ_HOST="localhost"
RABBITMQ_QUEUE="test_queue"
RABBITMQ_USER="guest"
RABBITMQ_PASS="guest"

echo "Consuming message from RabbitMQ..."
MESSAGE=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" -X POST \
    "http://$RABBITMQ_HOST:15672/api/queues/%2F/$RABBITMQ_QUEUE/get" \
    -H "Content-Type: application/json" \
    -d '{"count":1, "ackmode":"ack_requeue_false", "encoding":"auto", "truncate":500}')

MESSAGE_PAYLOAD=$(echo "$MESSAGE" | jq -r '.[0].payload')

if [ "$MESSAGE_PAYLOAD" != "null" ]; then
    echo "Received from RabbitMQ: $MESSAGE_PAYLOAD"
else
    echo "No messages in RabbitMQ queue."
fi