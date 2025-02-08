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

<img width="473" alt="Screenshot 2025-02-08 at 13 54 10" src="https://github.com/user-attachments/assets/fe6fb13a-531b-4364-8bcf-662c9da84ecb" />

Although the lru strategy attempts to select the first 2,000 keys, it is evident that the cache eviction strategy is stochastic.

<img width="492" alt="Screenshot 2025-02-08 at 14 00 33" src="https://github.com/user-attachments/assets/e6a8ff6e-4ae4-4d01-9b5f-a89d7a02c600" />


### Testing & Performance Evaluation

The project includes scripts to test each cache strategy using Siege for benchmarking:
	1.	Standard DB Query (0_simple_db_query.sh)
	2.	Unblocking Cache Fetch (1_unblocking_cache.sh)
	3.	External Blocking Recompute (2_external_blocking_cache.sh)
	4.	Probabilistic Expiration (4_probabilistic_expiration_cache.sh)

Each test:
	•	Sets a sample key in Redis.
	•	Runs a high-concurrency load test (`siege -c30 -t40S`).
	•	Deletes the key to ensure independent runs.

1. `0_simple_db_query.sh`

![Screenshot 2025-02-08 at 16 38 56](https://github.com/user-attachments/assets/061d786c-ed07-45b1-86a4-f7abe33b369b)

2. `1_unblocking_cache.sh`

![Screenshot 2025-02-08 at 16 44 18](https://github.com/user-attachments/assets/aaad6300-55c2-4aa2-92e2-10ed61f5c58c)

3. `2_external_blocking_cache.sh`

![Screenshot 2025-02-08 at 16 45 20](https://github.com/user-attachments/assets/92f0aac3-ffbb-4aca-8d11-7809d1077c7a)

4. `4_probabilistic_expiration_cache.sh`

![Screenshot 2025-02-08 at 16 47 15](https://github.com/user-attachments/assets/69c90b6e-b08c-4d49-b40e-e6e399bad4a8)
