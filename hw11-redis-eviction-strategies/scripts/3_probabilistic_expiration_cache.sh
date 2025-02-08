#!/bin/bash
echo "========================================"
echo "Running Test: Probabilistic Expiration"
echo "========================================"

curl -X POST "http://localhost:8080/set/samplekey/samplevalue/5"
echo -e "\nCache set successfully for samplekey"

echo -e "\nStarting siege load test..."
siege -c30 -t40S "http://localhost:8080/fetch/probabilistic/samplekey"

echo -e "\nCompleted test: Probabilistic Expiration"

echo -e "\nDeleting cache key..."
curl -X DELETE "http://localhost:8080/delete/samplekey"
echo -e "\nCache key deleted successfully!"