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

1. No index 
```
curl -I http://localhost:8080/measureSelectPerformance/
```
Results: 
![Screenshot 2025-01-25 at 20 01 42](https://github.com/user-attachments/assets/1b800797-b550-4bfc-b12b-ca27cd37157b)

2. BTREE index

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
![Screenshot 2025-01-25 at 19 28 04](https://github.com/user-attachments/assets/001f7907-e0bf-43ce-aa8d-5f6097e58425)

Run SELECT query with BTREE index:
```
curl -I http://localhost:8080/measureSelectPerformance/withBTREE
```

Results:
![Screenshot 2025-01-25 at 20 18 04](https://github.com/user-attachments/assets/2e27cd67-0703-4322-894b-9ce80e0177a1)

Remove index 
```
curl "http://localhost:8080/manageIndex?indexType=BTREE&action=delete"
```
Expected resonse: `BTREE index deleted successfully.`

3. HASH index

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

**Observations:**
1. First Query is Slow, Subsequent Queries are Faster:
Reasons:
- InnoDB Buffer Pool Caching. When we execute a query, MySQL reads the necessary rows from disk into memory (the InnoDB Buffer 	Pool) for the first query. For subsequent queries, the data is already loaded in memory (buffer pool), so no additional disk I/O is required, making them much faster.
- Even without indexes, MySQL may internally optimize and reorder queries when repeatedly executed. This is particularly true if the query is processed within the same connection/session.
3. HASH Index is Faster Than BTREE Index
Reasons:
- HASH indexes can still be faster than BTREE if the range is narrow, but they are generally not as efficient as BTREE for wide ranges.
- HASH indexes use a hash function to map the indexed column’s values directly to buckets. This allows near-instant lookup for matching rows in narrow ranges.

  

___

### Run Siege to simulate concurrent INSERT requests:
```
siege -c20 -t20S -f urls.txt
```
```
curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=0"
```
Expected resonse:  `innodb_flush_log_at_trx_commit set to 0`

siege -c20 -t20S -f urls.txt
img 30205    hits

curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=1"
innodb_flush_log_at_trx_commit set to 1

img 28790    hits

curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=2"
innodb_flush_log_at_trx_commit set to 2

img 30902    hits

siege -c40 -t20S -f urls.txt

innodb_flush_log_at_trx_commit set to 0
43027    hits


innodb_flush_log_at_trx_commit set to 1
img 25682    hits

innodb_flush_log_at_trx_commit set to 2
img 45236    hits



Expected Results

1. SELECT Performance
	•	Without Index: Queries are expected to be slower as the database performs a full table scan.
	•	With BTREE Index: Faster queries, as BTREE is optimized for range queries.
	•	With HASH Index: Performance may vary; HASH indexes are generally not optimized for range queries.

2. INSERT Performance
	•	innodb_flush_log_at_trx_commit = 0: Fastest inserts but with a risk of losing data in case of a crash.
	•	innodb_flush_log_at_trx_commit = 1: Slower inserts with full ACID compliance.
	•	innodb_flush_log_at_trx_commit = 2: Balance between speed and reliability.
