#!/bin/bash
echo "========================================"
echo "Running Test: Probabilistic Expiration"
echo "========================================"

# Set initial cache value before testing
curl -X POST "http://localhost:8080/set/samplekey/samplevalue/5"
echo -e "\nCache set successfully for samplekey"

# Run siege test
echo -e "\nStarting siege load test..."
siege -c30 -t40S "http://localhost:8080/fetch/probabilistic/samplekey"

echo -e "\nCompleted test: Probabilistic Expiration"

# Delete the cache key before finishing
echo -e "\nDeleting cache key..."
curl -X DELETE "http://localhost:8080/delete/samplekey"
echo -e "\nCache key deleted successfully!"