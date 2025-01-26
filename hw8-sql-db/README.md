## Project Overview

This project demonstrates performance testing for MySQL with InnoDB storage engine. 
It covers the following use cases:

1. Inserting a large dataset (40M users) into a MySQL table.
2. Measuring SELECT query performance with different indexing strategies:
   - No index
   - BTREE index
   - HASH index
3. Testing INSERT performance with different `innodb_flush_log_at_trx_commit` configurations.

App built using Go and MySQL env is set up with docker-compose.

### How to Use

**1. Start the Environment**

Run `docker-compose up -d` to spin up the project

**2. Build and run the Go app**

```
go build -o mysql-performance
./mysql-performance
```
Make POST request 
```
curl -X POST http://localhost:8080/insertUsers
```
*This might take up to 5 minutes to respond.*

![Screenshot 2025-01-25 at 19 58 36](https://github.com/user-attachments/assets/b30758ae-6584-4cd3-97a1-6604adf5bd01)

___

### SELECT quesries with different indexing strategies: 

**1. No index**
```
curl -I http://localhost:8080/measureSelectPerformance/
```

![Screenshot 2025-01-25 at 20 01 42](https://github.com/user-attachments/assets/1b800797-b550-4bfc-b12b-ca27cd37157b)

**2. BTREE index**

Creating index
```
curl "http://localhost:8080/manageIndex?indexType=BTREE&action=create"
```
Expected resonse: `BTREE index created successfully`

(Optional) Check that index has been created:
```
docker exec -it <container_name> mysql -u testuser -p
USE testdb;
SHOW INDEX FROM users;
```

![Screenshot 2025-01-25 at 20 13 40](https://github.com/user-attachments/assets/1a827146-8128-4424-a972-1ce3b6abb2fe)

Run SELECT query with BTREE index:
```
curl -I http://localhost:8080/measureSelectPerformance/withBTREE
```


![Screenshot 2025-01-25 at 20 18 04](https://github.com/user-attachments/assets/2e27cd67-0703-4322-894b-9ce80e0177a1)

Remove index 
```
curl "http://localhost:8080/manageIndex?indexType=BTREE&action=delete"
```
Expected resonse: `BTREE index deleted successfully.`

**3. HASH index**

Creating index
```
curl "http://localhost:8080/manageIndex?indexType=HASH&action=create"
```
Expected resonse: `HASH index created successfully.`

Run SELECT query with HASH index:
```
curl -I http://localhost:8080/measureSelectPerformance/withHASH
```
![Screenshot 2025-01-25 at 20 22 24](https://github.com/user-attachments/assets/b062cf28-cdac-4b2f-81d9-f12a6182308e)


Remove index 
```
curl "http://localhost:8080/manageIndex?indexType=HASH&action=delete"
```
Expected resonse: `HASH index deleted successfully.`

###Observations:###
**1. First Query is Slow, Subsequent Queries are Faster.**

Reasons:
- (With indexes) InnoDB Buffer Pool Caching. When we execute a query, MySQL reads the necessary rows from disk into memory (the InnoDB Buffer Pool) for the first query. For subsequent queries, the data is already loaded in memory (buffer pool), so no additional disk I/O is required, making them much faster.
- (No indexes) Even without indexes, MySQL may internally optimize and reorder queries when repeatedly executed. This is particularly true if the query is processed within the same connection/session.

**2. HASH Index is Faster Than BTREE Index**

This outcome was unexpected, given that BTREE indexes are specifically optimized for range queries.
The observed performance advantage of HASH indexes over BTREE is nuanced and requires deeper exploration.

*__Important note:__ MySQL’s InnoDB engine doesn’t support native HASH indexes, created HASH index is still implemented as Index_type BTREE under the hood.*

Results from EXPLAIN ANALYZE show:
BTREE Index (idx_dob_btree):
   - Time: 16ms to 185ms
HASH Index (idx_dob_hash):
   - Time: 2.65ms to 33.8ms

img

Both indexes have the same cardinality and range, and identical query structures were used. Yet, the “HASH” index consistently performs faster.

**Potential Causes**
The faster performance of the “HASH” index may be attributed to the following:
	1.	Query Optimizer Behavior: MySQL may assign different cost estimations or prioritize execution plans for the “HASH” index, even though it is implemented as a BTREE.
	2.	Subtle Metadata Differences: The USING HASH keyword could influence MySQL’s internal handling of the index, leading to optimizations such as prefetching or read-ahead operations.
	3.	Range Handling: HASH-like indexing may provide advantages for narrow range lookups due to differences in query planning, even though it is not designed for wide ranges.


We can try to set profiling for booth quesries 
```
SET PROFILING = 1;
SELECT * FROM users FORCE INDEX (idx_dob_btree) WHERE date_of_birth BETWEEN '1990-01-01' AND '1991-01-01' LIMIT 1000;
SELECT * FROM users FORCE INDEX (idx_dob_hash) WHERE date_of_birth BETWEEN '1990-01-01' AND '1991-01-01' LIMIT 1000;

SHOW PROFILES;
SHOW PROFILE ALL FOR QUERY 1;
SHOW PROFILE ALL FOR QUERY 2;
```

Profiling Results

The profiling tests for both queries (BTREE and HASH indices) reveal negligible differences in CPU usage and overall duration. While the “HASH” index shows slightly better profiling metrics, the observed performance differences cannot be fully explained by these numbers alone.


Set profiling for adjusted date range and query’s modified LIMIT. 

img 1 
img 2
img 3

Scenario 1: RANGE 1980-2015 LIMIT 12,000
1.	HASH Index (idx_dob_hash)
   - Actual time: 0.764..1.548 seconds 
2.	BTREE Index (idx_dob_btree)
   - Actual time: 1.06..1.21 seconds 
Scenario 2: RANGE 1980-2017 LIMIT 50,000
1.	HASH Index (idx_dob_hash)
   - Actual time: 0.787..5.395 seconds
2.	BTREE Index (idx_dob_btree)
   - Actual time: 0.953..0.669 seconds
Scenario 3: RANGE 1980-2018 LIMIT 100,000
1.	HASH Index (idx_dob_hash)
   - Actual time: 0.995..10.585 seconds 
2.	BTREE Index (idx_dob_btree)
   - Actual time: 2.65..11.101 seconds 

Summary of Differences:
- The HASH index is optimized for smaller RANGE and lower LIMIT values, resulting in better performance in such scenarios.
- The BTREE index performs more efficiently as the RANGE and LIMIT size increase, with the performance advantage of the HASH index diminishing and eventually reversing.
- The observed differences arise from how MySQL optimizes queries for the HASH index compared to the BTREE index, rather than any inherent differences in the underlying index structures (as both are implemented as BTREE).

___

### Testing INSERT performance

**Run Siege to simulate concurrent INSERT requests:**

```
siege -c20 -t20S -f urls.txt
```
```
curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=0"
```
Expected resonse:  `innodb_flush_log_at_trx_commit set to 0`

```
siege -c20 -t20S -f urls.txt
```
<img width="437" alt="Screenshot 2025-01-25 at 20 32 16" src="https://github.com/user-attachments/assets/e0909e6e-0823-4245-8c18-e9bd845b126f" />

```
curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=1"
```
Expected resonse:  `innodb_flush_log_at_trx_commit set to 1`

<img width="437" alt="Screenshot 2025-01-25 at 20 33 49" src="https://github.com/user-attachments/assets/84886eb7-9de6-409b-be62-cd27661291ea" />

```
curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=2"
```
Expected resonse:  `innodb_flush_log_at_trx_commit set to 2`

<img width="432" alt="Screenshot 2025-01-25 at 20 34 58" src="https://github.com/user-attachments/assets/3ee8cfa1-db73-4890-9790-29451aeff3e8" />



```
siege -c40 -t20S -f urls.txt
```

Results for `innodb_flush_log_at_trx_commit set to 0`

<img width="439" alt="Screenshot 2025-01-25 at 20 36 28" src="https://github.com/user-attachments/assets/3fa1f05d-d908-49ac-8d85-051e0762410a" />

Results for `innodb_flush_log_at_trx_commit set to 1`

<img width="437" alt="Screenshot 2025-01-25 at 20 37 34" src="https://github.com/user-attachments/assets/7da1aea8-2100-4f21-9619-416fa0b15912" />


Results for `innodb_flush_log_at_trx_commit set to 2`

<img width="433" alt="Screenshot 2025-01-25 at 20 38 39" src="https://github.com/user-attachments/assets/f4b92934-8faf-473f-b5d7-7c6e4e35bb43" />


Quick Overview:
- Setting innodb_flush_log_at_trx_commit = 0 focuses on performance by skipping log flushes to disk after each transaction, which increases the risk of data loss in the event of a crash.
- Setting innodb_flush_log_at_trx_commit = 1 prioritizes durability by flushing the log after every commit, but this significantly reduces performance.
- Setting innodb_flush_log_at_trx_commit = 2 provides a balanced approach, flushing the log less frequently (once per second), offering a compromise between performance and durability.

As the number of concurrent users increases (from 20 to 40), the advantages of innodb_flush_log_at_trx_commit = 2 become more apparent. It delivers performance comparable to trx_commit set to `0` while maintaining better transaction durability.