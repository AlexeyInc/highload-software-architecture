#!/bin/bash
 
container_name="apache_benchmark"
 
if ! docker ps --format '{{.Names}}' | grep -q "^$container_name$"; then
  echo "Error: Container $container_name is not running."
  echo "Please start the container using: docker-compose up"
  exit 1
fi
 
ab_command="ab -n 10000 -c 50 http://app:8081/"
 
echo "Running stress test: $ab_command"
docker exec -it $container_name sh -c "$ab_command"

echo "Stress test completed."