## Project Overview

This project demonstrates performance testing for MySQL with InnoDB storage engine. 
It covers the following use cases:

1. Inserting a large dataset (40M users) into a MySQL table.
2. Measuring SELECT query performance with different indexing strategies:
   - No index
   - BTREE index
   - HASH index
3. Testing INSERT performance with different `innodb_flush_log_at_trx_commit` configurations.

The program is built using Go, and the (MySQL env is set up with docker-compose).

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
(Might take up to 5 minutes to respond)


curl "http://localhost:8080/manageIndex?indexType=BTREE&action=create"
BTREE index created successfully

curl "http://localhost:8080/manageIndex?indexType=HASH&action=delete"

Optional:
docker exec -it <container_name> mysql -u testuser -p
USE testdb;
SHOW INDEX FROM users;

curl -I http://localhost:8080/measureSelectPerformance/withBTREE

img 99.32975ms

curl "http://localhost:8080/manageIndex?indexType=BTREE&action=delete"
BTREE index deleted successfully.%   

curl "http://localhost:8080/manageIndex?indexType=HASH&action=create" 
HASH index created successfully.% 

img 32.968667ms

curl "http://localhost:8080/manageIndex?indexType=HASH&action=delete" 
HASH index deleted successfully.%   

-----

Run Siege to simulate concurrent requests:
siege -c20 -t20S -f urls.txt

curl -X POST "http://localhost:8080/changeFlushLogSetting?innodb_flush_log_at_trx_commit=0"
innodb_flush_log_at_trx_commit set to 0

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