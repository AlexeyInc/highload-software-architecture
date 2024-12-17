#!/bin/bash
 
if [ "$#" -lt 2 ]; then
  echo "$0 -n <number_of_requests> -c <concurrent_requests>"
  exit 1
fi
 
while [[ "$#" -gt 0 ]]; do
  case $1 in
    -n) num_requests="$2"; shift ;;
    -c) concurrency="$2"; shift ;;
    *) echo "Unknown param: $1"; exit 1 ;;
  esac
  shift
done

container_name="apache_benchmark"
 
if ! docker ps --format '{{.Names}}' | grep -q "^$container_name$"; then
  echo "Error: Container $container_name is not running."
  echo "Please start the container using: docker-compose up"
  exit 1
fi
 
ab_command="ab -n $num_requests -c $concurrency http://host.docker.internal:8080/"

echo "Running stress test: $ab_command"
docker exec -it $container_name sh -c "$ab_command"

echo "Stress test completed."

# ./run_ab_test.sh -n 10000 -c 50