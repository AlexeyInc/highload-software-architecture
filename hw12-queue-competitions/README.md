## Project Overview

This project evaluates RabbitMQ, Redis RDB, and Redis AOF as message queues by testing their throughput, latency, and scalability under high load.

---

### Key Features

1. Implemented HTTP endpoints for publishing messages:

- `/publish/rabbitmq`
- `/publish/redis_rdb`
- `/publish/redis_aof`
- `/publish/:queue_name → Dynamic queue selection`

2. Background Consumers for Real-Time Processing
- RabbitMQ → Listens via AMQP (ConsumeRabbitMQ)
- Redis RDB → Uses BLPOP to process messages (ConsumeRedisRDB)
- Redis AOF → Uses BLPOP (ConsumeRedisAOF)

3. High-Load Testing with siege
- Simulated 25, 50, 100 concurrency
- Measured TPS (transactions per second), response time, and latency

4. Monitoring & Metrics
- TIG staack for RabbitMQ
- Grafana + Prometheus for redis

5. `scripts` folder contains essential bash scripts to manually push/pull messages from different queues

These scripts are useful for:
- Manually verifying queue functionality before running load tests
- Debugging message flow and ensuring correct processing
- Checking message persistence & retrieval behavior
- Ensuring messages are consumed properly

```
./push_rabbitmq.sh  # Sends a message to RabbitMQ
./push_redis_rdb.sh  # Sends a message to Redis RDB
./push_redis_aof.sh  # Sends a message to Redis AOF
/pull_rabbitmq.sh  # Retrieves and removes a message from RabbitMQ
./pull_redis_rdb.sh  # Retrieves and removes a message from Redis RDB
./pull_redis_aof.sh  # Retrieves and removes a message from Redis AOF
```
For add message to the queue and verify that message are being processed correctly.

```
./generic_push.sh rabbitmq  # Push message to RabbitMQ
./generic_pull.sh redis_rdb  # Pull message from Redis RDB
```
Saves time by allowing a single script for multiple queues.

`messages_started.lock`
- Purpose: A lock file used to ensure that consumers only start when messages exist.
- Usage: Created when pushing messages and deleted when pulling messages.

---

## Results

**RabbitMQ**

1. `siege -c25 -t30S -b "http://localhost:8080/publish/rabbitmq"`

```
Lifting the server siege...
Transactions:		   168636    hits
Availability:		      100.00 %
Elapsed time:		       30.53 secs
Data transferred:	        7.08 MB
Response time:		        4.49 ms
Transaction rate:	     5523.62 trans/sec
Throughput:		        0.23 MB/sec
Concurrency:		       24.80
Successful transactions:   168636
Failed transactions:	        0
Longest transaction:	      110.00 ms
Shortest transaction:	        0.00 ms
```
`curl "http://localhost:8080/msg/counter/rabbitmq"`

`Message consumed in rabbitmq: 168636`

2. `siege -c50 -t30S -b "http://localhost:8080/publish/rabbitmq"`
```
Lifting the server siege...
Transactions:		   172434    hits
Availability:		      100.00 %
Elapsed time:		       30.79 secs
Data transferred:	        7.24 MB
Response time:		        8.89 ms
Transaction rate:	     5600.32 trans/sec
Throughput:		        0.23 MB/sec
Concurrency:		       49.79
Successful transactions:   172434
Failed transactions:	        0
Longest transaction:	      110.00 ms
Shortest transaction:	        0.00 ms
```
`Message consumed in rabbitmq: 172434`

3. `siege -c100 -t30S -b "http://localhost:8080/publish/rabbitmq"`
```
Lifting the server siege...
Transactions:		   143864    hits
Availability:		      100.00 %
Elapsed time:		       30.27 secs
Data transferred:	        6.04 MB
Response time:		       20.98 ms
Transaction rate:	     4752.69 trans/sec
Throughput:		        0.20 MB/sec
Concurrency:		       99.73
Successful transactions:   143864
Failed transactions:	        0
Longest transaction:	      330.00 ms
Shortest transaction:	        0.00 ms
```
`Message consumed in rabbitmq: 143864`


Concurency 25             |  Concurency 50
:-------------------------:|:-------------------------:
![Screenshot 2025-02-15 at 15 52 14](https://github.com/user-attachments/assets/cb744976-8ae6-4c8e-9cc9-3e804b97ce08)  |  ![Screenshot 2025-02-15 at 15 52 19](https://github.com/user-attachments/assets/c96ae4ad-c18a-41b3-8fad-9ddba9961f10)

At all concurrency levels, the peak number of published but unconsumed messages reached 65K, regardless of whether the concurrency was 25, 50 or 100.

----

**Redis RDB**

1. `siege -c25 -t30S -b "http://localhost:8080/publish/redis_rdb"`
```
Lifting the server siege...
Transactions:		   174962    hits
Availability:		      100.00 %
Elapsed time:		       30.49 secs
Data transferred:	        5.67 MB
Response time:		        4.31 ms
Transaction rate:	     5738.34 trans/sec
Throughput:		        0.19 MB/sec
Concurrency:		       24.76
Successful transactions:   174963
Failed transactions:	        0
Longest transaction:	      180.00 ms
Shortest transaction:	        0.00 ms
```
`curl "http://localhost:8080/msg/counter/redis_rdb"`

`Message consumed in redis_rdb: 91481`

~1-2s latter
`Message consumed in redis_rdb: 174962`


2. `siege -c50 -t30S -b "http://localhost:8080/publish/redis_rdb"`
```
Lifting the server siege...
Transactions:		   176275    hits
Availability:		      100.00 %
Elapsed time:		       30.69 secs
Data transferred:	        7.40 MB
Response time:		        8.67 ms
Transaction rate:	     5743.73 trans/sec
Throughput:		        0.24 MB/sec
Concurrency:		       49.79
Successful transactions:   176275
Failed transactions:	        0
Longest transaction:	      220.00 ms
Shortest transaction:	        0.00 ms
```

`Message consumed in redis_rdb: 101506`

~1-2s latter
`Message consumed in redis_rdb: 176275`

3. `siege -c100 -t30S -b "http://localhost:8080/publish/redis_rdb"`
```
Lifting the server siege...
Transactions:		   167585    hits
Availability:		      100.00 %
Elapsed time:		       30.54 secs
Data transferred:	        7.03 MB
Response time:		       18.17 ms
Transaction rate:	     5487.39 trans/sec
Throughput:		        0.23 MB/sec
Concurrency:		       99.71
Successful transactions:   167585
Failed transactions:	        0
Longest transaction:	     1040.00 ms
Shortest transaction:	        0.00 ms
```
`Message consumed in redis_rdb: 100966`

~1-2s latter
`Message consumed in redis_rdb: 167585`

-----

**Redis AOF**

1. `siege -c25 -t30S -b "http://localhost:8080/publish/redis_aof"`
```
Lifting the server siege...
Transactions:		   178606    hits
Availability:		      100.00 %
Elapsed time:		       30.57 secs
Data transferred:	        7.49 MB
Response time:		        4.24 ms
Transaction rate:	     5842.53 trans/sec
Throughput:		        0.25 MB/sec
Concurrency:		       24.79
Successful transactions:   178606
Failed transactions:	        0
Longest transaction:	       60.00 ms
Shortest transaction:	        0.00 ms
```
`curl "http://localhost:8080/msg/counter/redis_aof"`

`"Message consumed in redis_aof: 92192"`

~1-2s latter
`Message consumed in redis_aof: 178606`

2. `siege -c50 -t30S -b "http://localhost:8080/publish/redis_aof"`
```
Lifting the server siege...
Transactions:		   172954    hits
Availability:		      100.00 %
Elapsed time:		       30.33 secs
Data transferred:	        7.26 MB
Response time:		        8.73 ms
Transaction rate:	     5702.41 trans/sec
Throughput:		        0.24 MB/sec
Concurrency:		       49.76
Successful transactions:   172954
Failed transactions:	        0
Longest transaction:	      210.00 ms
Shortest transaction:	        0.00 ms
```
`Message consumed in redis_aof: 93101`

~1-2s latter
`Message consumed in redis_aof: 172954`

3. `siege -c100 -t30S -b "http://localhost:8080/publish/redis_aof"`
```
Lifting the server siege...
Transactions:		   164058    hits
Availability:		      100.00 %
Elapsed time:		       30.72 secs
Data transferred:	        6.88 MB
Response time:		       18.67 ms
Transaction rate:	     5340.43 trans/sec
Throughput:		        0.22 MB/sec
Concurrency:		       99.69
Successful transactions:   164058
Failed transactions:	        0
Longest transaction:	      250.00 ms
Shortest transaction:	        0.00 ms
```
`Message consumed in redis_aof: 95806`

~1-2s latter
`Message consumed in redis_aof: 164058`

---

## Final thoughts

1. **Best for High Message Throughput** -> **Redis AOF & Redis RDB**
- Redis (AOF & RDB) outperforms RabbitMQ in throughput, handling ~5,700-5,800 TPS.

2. **Best for nearly Real-Time Processing** -> **RabbitMQ**
- RabbitMQ instantly consumes messages, while Redis queues show a 1-2s delay.
- Redis BLPOP is slower at high loads, affecting real-time applications.

3. **Large-Scale Parallel Consumers** -> Depends on the use case
- RabbitMQ scales well, but at 100 concurrent clients, throughput drops by ~15%.
- Redis RDB & AOF also drop at high concurrency but maintain higher TPS.

4. **Best for Durable Queue Persistence** -> Depends on failure tolerance & performance needs
- **Redis AOF** (Append-Only File) persists every operation, ensuring data safety.
- **RabbitMQ** also guarantees durability but has more overhead.
Assumtions:
- Redis AOF if data safety is more important than performance.
- RabbitMQ if you need persistent, transactional message processing with requeuing capabilities.

### Insights:
**Disk Write Efficiency:**
Redis RDB slightly outperforms AOF in disk write efficiency when saving snapshots less frequently.
- In tests, RDB saved data every 1 second when at least one change occurred, which resulted in lower disk I/O overhead compared to continuous writes in AOF.
- If write efficiency is a concern, RDB can be a viable alternative to AOF when some data loss is acceptable.
