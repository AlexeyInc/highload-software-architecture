#!/bin/bash

# RabbitMQ settings
RABBITMQ_HOST="localhost"
RABBITMQ_QUEUE="test_queue"
RABBITMQ_EXCHANGE="test_exchange"
RABBITMQ_ROUTING_KEY="test_key"

# Define message payload
MESSAGE="Hello, RabbitMQ!"

echo "Setting up RabbitMQ queue and exchange..."
EXCHANGE_EXISTS=$(docker exec rabbitmq rabbitmqadmin list exchanges name | grep -w $RABBITMQ_EXCHANGE)

if [ -z "$EXCHANGE_EXISTS" ]; then
    echo "Declaring RabbitMQ exchange..."
    docker exec rabbitmq rabbitmqadmin declare exchange name=$RABBITMQ_EXCHANGE type=direct
    docker exec rabbitmq rabbitmqadmin declare queue name=$RABBITMQ_QUEUE durable=true
    docker exec rabbitmq rabbitmqadmin declare binding source=$RABBITMQ_EXCHANGE destination_type=queue destination=$RABBITMQ_QUEUE routing_key=$RABBITMQ_ROUTING_KEY
fi

echo "Publishing message to RabbitMQ..."
docker exec rabbitmq rabbitmqadmin publish exchange=$RABBITMQ_EXCHANGE routing_key=$RABBITMQ_ROUTING_KEY payload="$MESSAGE"

echo "Message sent to RabbitMQ!"