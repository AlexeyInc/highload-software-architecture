#!/bin/bash

QUEUE_TYPE=$1  # "rabbitmq", "redis_rdb", or "redis_aof"

# RABBITMQ data
RABBITMQ_HOST="localhost"
RABBITMQ_QUEUE="test_queue"
RABBITMQ_EXCHANGE="test_exchange"
RABBITMQ_ROUTING_KEY="test_key"

# payload message
MESSAGE="Hello, queue!"

# lock file to signal consumers
touch ./messages_started.lock


EXCHANGE_EXISTS=$(docker exec rabbitmq rabbitmqadmin list exchanges name | grep -w $RABBITMQ_EXCHANGE)

if [ -z "$EXCHANGE_EXISTS" ]; then
    echo "Setting up RabbitMQ queue and exchange..."
    docker exec rabbitmq rabbitmqadmin declare exchange name=$RABBITMQ_EXCHANGE type=direct
    docker exec rabbitmq rabbitmqadmin declare queue name=$RABBITMQ_QUEUE durable=true
    docker exec rabbitmq rabbitmqadmin declare binding source=$RABBITMQ_EXCHANGE destination_type=queue destination=$RABBITMQ_QUEUE routing_key=$RABBITMQ_ROUTING_KEY
fi

echo "Pushing message to $QUEUE_TYPE..."
# choose queue to push
case $QUEUE_TYPE in
    "rabbitmq")
        docker exec rabbitmq rabbitmqadmin publish exchange=test_exchange routing_key=test_key payload="$MESSAGE"
        ;;
    "redis_rdb")
        docker exec redis_rdb redis-cli RPUSH test_queue "$MESSAGE"
        ;;
    "redis_aof")
        docker exec redis_aof redis-cli RPUSH test_queue "$MESSAGE"
        ;;
    *)
        echo "Invalid queue type. Use 'rabbitmq', 'redis_rdb', or 'redis_aof'."
        exit 1
        ;;
esac

echo "Message sent to $QUEUE_TYPE!"