## Project Overview

This project implements a Redis Cluster with master-slave replication, supports various eviction strategies, and integrates a probabilistic early expiration mechanism to prevent cache stampedes. The system provides multiple cache interaction strategies through an HTTP API.

### Key Features
1. Redis Cluster with Sentinel & Eviction Strategies
- A Redis Cluster with 6 nodes is set up using Docker Compose.
- Supports different eviction policies, including:
- allkeys-lru
- allkeys-lfu
- allkeys-random
- volatile-ttl
- A monitoring service listens for key evictions.
2. Cache Interaction Strategies
- Direct Database Fetch (/fetch/db/{key}): Retrieves data directly from a mock DB.
- Basic Caching (/fetch/cache/{key}): Fetches from cache, recomputes on cache miss.
- External Recompute with Locking (/fetch/external/{key}): Ensures recomputation occurs only once using locks.
- Probabilistic Early Expiration (/fetch/probabilistic/{key}): Implements a probabilistic cache clearing approach using:

\text{probability} = 1 - e^{-\beta(1 - \text{remainingRatio})}

This method spreads out recomputation probabilistically rather than allowing all requests to trigger a refresh at once.

3. Monitoring and Benchmarking
- RedisInsight: Provides a UI for visualizing keys, memory, and cluster nodes.
- Prometheus & Redis Exporter: Collects Redis performance metrics.
- Grafana: Displays real-time cache analytics.
- Siege Load Testing: Simulates concurrent requests to test cache efficiency.

### How to Use

1. Start Redis Cluster & Services `docker-compose up -d`. 
Check connection to redis cluster via redisinsight on `http://localhost:5540`
(additional info available in grafana on `http://localhost:3000`)

2. Run Load Tests
```
chmod +x *.sh
./0_simple_db_query.sh
./1_unblocking_cache.sh
./2_external_blocking_cache.sh
./4_probabilistic_expiration_cache.sh
```

3. Monitor Metrics
- RedisInsight: http://localhost:5540
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090

### Cache eviction

Run `curl "http://localhost:8080/preloadKeys"` to init 10000 keys. 



### Testing & Performance Evaluation

The project includes scripts to test each cache strategy using Siege for benchmarking:
	1.	Standard DB Query (0_simple_db_query.sh)
	2.	Unblocking Cache Fetch (1_unblocking_cache.sh)
	3.	External Blocking Recompute (2_external_blocking_cache.sh)
	4.	Probabilistic Expiration (4_probabilistic_expiration_cache.sh)

Each test:
	•	Sets a sample key in Redis.
	•	Runs a high-concurrency load test (siege -c30 -t40S).
	•	Deletes the key to ensure independent runs.