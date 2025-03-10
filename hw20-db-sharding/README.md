## Project overview

Project explores horizontal sharding strategies in PostgreSQL by comparing three approaches: no sharding, FDW (Foreign Data Wrapper), and Citus.  Go application inserts and reads 1 million book records, benchmarking performance across these methods. The FDW approach partitions data across `postgresql-b1` and `postgresql-b2` using foreign tables and a unified view, while Citus dynamically distributes data across `postgresql-b1` and `postgresql-b2` based on hash-based sharding. Performance tests using siege helped evaluate insert and read efficiency for each approach. The results provide insights into the trade-offs between simplicity (no sharding), FDW’s manual partitioning, and Citus’ automated scaling capabilities.

## Configure sharding with FDW (Foreign Data Wrapper)

*Instructions to configure sharding using FDW (Foreign Data Wrapper) and Citus in a PostgreSQL cluster.*

**Step 1: Set up tables on each shard**

Ensure you have run: `docker-compose.fdw.yaml`

Create tables on worker nodes (postgresql-b1 and postgresql-b2) to store specific category_id values.

On `postgresql-b1` (Shard 1)

Run:
```
docker exec -it postgresql-b1 psql -U postgres -d books_1
```
Create table & index:
```
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id INT NOT NULL,
    CONSTRAINT category_id_check CHECK ( category_id = 1 OR category_id = 2 ),
    author VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    year INT NOT NULL
); 

CREATE INDEX books_category_id_idx ON books USING btree(category_id);
```

On `postgresql-b2` (Shard 2)

Run:
```
docker exec -it postgresql-b2 psql -U postgres -d books_2
```
Create table & index:
```
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id INT NOT NULL, 
    author VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    year INT NOT NULL
);

CREATE INDEX books_category_id_idx ON books USING btree(category_id);
```

**Step 2: Configure FDW on Main DB (postgresql-b)**

Connect to the main coordinator node:
```
docker exec -it postgresql-b psql -U postgres -d books_db
```
Enable FDW extension:
```
CREATE EXTENSION postgres_fdw;
```
Add foreign servers
```
CREATE SERVER books_1_server 
FOREIGN DATA WRAPPER postgres_fdw 
OPTIONS (host 'postgresql-b1', port '5432', dbname 'books_1');

CREATE SERVER books_2_server 
FOREIGN DATA WRAPPER postgres_fdw 
OPTIONS (host 'postgresql-b2', port '5432', dbname 'books_2');
```

Create user mappings:
```
CREATE USER MAPPING FOR postgres
SERVER books_1_server
OPTIONS (user 'postgres', password 'postgres');

CREATE USER MAPPING FOR postgres
SERVER books_2_server
OPTIONS (user 'postgres', password 'postgres');
```

Create foreign tables on `postgresql-b`
```
CREATE FOREIGN TABLE books_1 (
    id UUID DEFAULT gen_random_uuid(),
    category_id INT NOT NULL,
    author VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    year INT NOT NULL
) SERVER books_1_server
OPTIONS (schema_name 'public', table_name 'books');

CREATE FOREIGN TABLE books_2 (
    id UUID DEFAULT gen_random_uuid(),
    category_id INT NOT NULL,
    author VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    year INT NOT NULL
) SERVER books_2_server
OPTIONS (schema_name 'public', table_name 'books');
```

**Step 3: Replace books table with a view**

Since data is stored in separate foreign tables, create a view that combines both:
```
DROP TABLE IF EXISTS books;

CREATE VIEW books AS
    SELECT * FROM books_1
    UNION ALL
    SELECT * FROM books_2;
```

**Step 4: Set Up Insert Rules**

To ensure that new records go into the correct shard, set up insert rules:

```
CREATE RULE books_insert_to_1 AS ON INSERT TO books
WHERE (category_id = 1 OR category_id = 2)
DO INSTEAD INSERT INTO books_1 VALUES (NEW.*);

CREATE RULE books_insert_to_2 AS ON INSERT TO books
WHERE (category_id = 3 OR category_id = 4)
DO INSTEAD INSERT INTO books_2 VALUES (NEW.*);
```

To prevent unintended inserts, updates, or deletes, add rules to block direct modifications:

```
CREATE RULE books_insert AS ON INSERT TO books DO INSTEAD NOTHING;
CREATE RULE books_update AS ON UPDATE TO books DO INSTEAD NOTHING;
CREATE RULE books_delete AS ON DELETE TO books DO INSTEAD NOTHING;
```


## Configure sharding with Citus

Ensure you have run: `docker-compose.citus.yaml`

**Step 1: Start Citus and set the coordinator**

Run:
```
docker exec -it postgresql-b psql -U postgres -d books_db
```
```
SELECT citus_set_coordinator_host('postgresql-b', 5432);
```

**Step 2: Register worker nodes (`postgresql-b1` & `postgresql-b2`)**

```
SELECT citus_add_node('postgresql-b1', 5432);
SELECT citus_add_node('postgresql-b2', 5432);
```

Confirm the worker nodes are properly registered:
```
SELECT * FROM citus_get_active_worker_nodes();
```

Expected output:
```
   node_name   | node_port 
---------------+-----------
 postgresql-b2 |      5432
 postgresql-b1 |      5432
 ```

**Step 3: Create Distributed Table on `postgresql-b`**
```
CREATE TABLE books (
    id UUID DEFAULT gen_random_uuid(),
    category_id INT NOT NULL,
    author VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    year INT NOT NULL,
    PRIMARY KEY (category_id, id)
);
```

Distribute the table across shards (`postgresql-b1` & `postgresql-b2`):

```
SELECT create_distributed_table('books', 'category_id', colocate_with => 'none', shard_count => 2);
```

**Step 4: Rebalance the shards**
```
SELECT rebalance_table_shards();
```

Check shard performance: `EXPLAIN ANALYZE SELECT * FROM books;`

Check all distributed tables: `SELECT * FROM citus_tables;`


## (Without sharding)

The setup will be configured automatically when launching main.go, and all commands can be executed seamlessly.
 

## Results


**Single postgres node**

```
curl -X GET http://localhost:8080/insert
Inserted 1000000 records in 16.030161458s
```
```
siege -c50 -t20S "http://localhost:8080/read"

Lifting the server siege...
Transactions:		     1667    hits
Availability:		      100.00 %
Elapsed time:		       20.29 secs
Data transferred:	        1.49 MB
Response time:		      596.77 ms
Transaction rate:	       82.16 trans/sec
Throughput:		        0.07 MB/sec
Concurrency:		       49.03
Successful transactions:     1667
Failed transactions:	        0
Longest transaction:	     1820.00 ms
Shortest transaction:	      110.00 ms
```
```
siege -c50 -t20S "http://localhost:8080/insert-batch"

Lifting the server siege...
Transactions:		    22739    hits
Availability:		      100.00 %
Elapsed time:		       21.00 secs
Data transferred:	        0.75 MB
Response time:		       35.34 ms
Transaction rate:	     1082.81 trans/sec
Throughput:		        0.04 MB/sec
Concurrency:		       38.27
Successful transactions:    22739
Failed transactions:	        0
Longest transaction:	     4550.00 ms
Shortest transaction:	        0.00 ms
```

**FDW**

```
curl -X GET http://localhost:8080/insert
Inserted 1000000 records in 37.825414292s
```
```
siege -c50 -t20S "http://localhost:8080/read"

Lifting the server siege...
Transactions:		      596    hits
Availability:		      100.00 %
Elapsed time:		       20.36 secs
Data transferred:	       68.10 MB
Response time:		     1660.18 ms
Transaction rate:	       29.27 trans/sec
Throughput:		        3.34 MB/sec
Concurrency:		       48.60
Successful transactions:      596
Failed transactions:	        0
Longest transaction:	     6180.00 ms
Shortest transaction:	      200.00 ms
```
```
siege -c50 -t20S "http://localhost:8080/insert-batch"

Lifting the server siege...
Transactions:		    11288    hits
Availability:		      100.00 %
Elapsed time:		       20.28 secs
Data transferred:	        0.38 MB
Response time:		       89.43 ms
Transaction rate:	      556.61 trans/sec
Throughput:		        0.02 MB/sec
Concurrency:		       49.78
Successful transactions:    11288
Failed transactions:	        0
Longest transaction:	     4030.00 ms
Shortest transaction:	        0.00 ms
```

**Citus**
```
curl -X GET http://localhost:8080/insert
Inserted 1000000 records in 34.830897167s
```
```
siege -c50 -t20S "http://localhost:8080/read"

Lifting the server siege...
Transactions:		     1848    hits
Availability:		      100.00 %
Elapsed time:		       20.44 secs
Data transferred:	      212.26 MB
Response time:		      543.38 ms
Transaction rate:	       90.41 trans/sec
Throughput:		       10.38 MB/sec
Concurrency:		       49.13
Successful transactions:     1848
Failed transactions:	        0
Longest transaction:	     3300.00 ms
Shortest transaction:	       90.00 ms
```
```
siege -c50 -t20S "http://localhost:8080/insert-batch"

Lifting the server siege...
Transactions:		     3701    hits
Availability:		      100.00 %
Elapsed time:		       20.34 secs
Data transferred:	        0.12 MB
Response time:		      271.70 ms
Transaction rate:	      181.96 trans/sec
Throughput:		        0.01 MB/sec
Concurrency:		       49.44
Successful transactions:     3701
Failed transactions:	        0
Longest transaction:	     2260.00 ms
Shortest transaction:	        0.00 ms
```

**Summary:**
- Sharding overhead is expected, as all shards were on the same machine, limiting real-world benefits like parallelism and load distribution. 
- Citus performance could improve with better coordinator-worker balance and worker-specific tuning.
- FDW is best when you want manual shard control (e.g., category-based placement).
- Citus is best for high-performance, large-scale sharding with automated balancing.