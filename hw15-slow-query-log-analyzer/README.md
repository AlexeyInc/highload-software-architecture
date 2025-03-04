## Project Overview

This project sets up a MySQL database with slow query logging, ELK (Elasticsearch, Logstash, Kibana), and Graylog to compare logging performance under different MySQL `long_query_time` values. The application is Go based API with endpoints for simulating slow queries and searching users.

## How to use

**Database Initialization**

*During application startup, the system automatically inserts 1 million records into the MySQL database. This ensures that the `/search` endpoint has data available for queries and performance testing.*


**1. To start all services (mysql, Go app, ELK, Graylog, Filebeat and Logstash), run:**

`docker-compose up --build`

**2. After starting the services, configure Graylog to receive MySQL slow query logs**
```
docker cp ./scripts/graylog_setup_beats.sh graylog:/tmp/

docker exec -it graylog bash -c "bash /tmp/graylog_setup_beats.sh"
```
Expected Output:
`{"id":"your_id"}Beats Input has been successfully created!`

**3.  Verify logs in Kibana & Graylog**

Make test reqeust multiple times: `curl http://localhost:8080/slow?timeout=1` and check logs:

Kibana (`http://localhost:5601`): Navigate to "Discover", search for logs in the `mysql-slow-logs-*` index.

![Screenshot 2025-03-04 at 12 46 45](https://github.com/user-attachments/assets/ec3e2f92-a8e7-4e7d-a7d8-98ae0f144c03)

</br>

Graylog (`http://localhost:9000`): Log in (default `admin:admin`) and go to search page.

![Screenshot 2025-03-04 at 12 59 55](https://github.com/user-attachments/assets/139efdcc-1f6a-4125-b217-95ee033c54be)

____

## Running performance tests

Changing `long_query_time` and Run Siege

**Test 1**

`long_query_time = 0` (Log Everything)
```
    docker exec -it mysql-db mysql -uroot -prootpassword -e "SET GLOBAL long_query_time = 0;"

    docker exec -it mysql-db mysql -uroot -prootpassword -e "SHOW VARIABLES LIKE 'long_query_time';"
```

Expected output:
```
    +-----------------+----------+
    | Variable_name   | Value    |
    +-----------------+----------+
    | long_query_time | 0.000000 |
    +-----------------+----------+
```

Run search query stress test:
    `siege -c30 -t30S "http://localhost:8080/search?name=User50"`

Results:
```
    Data transferred:  892.92 MB
    Throughput:       29.70 MB/sec
    Successful transactions: 1264
    Longest transaction:  8670.00 ms
    Shortest transaction:  160.00 ms
```
**Test 2**

`long_query_time = 0.3` (Log Queries > 300ms)

`docker exec -it mysql-db mysql -uroot -prootpassword -e "SET GLOBAL long_query_time = 0.3;"`

Results:
```
    Data transferred:  998.88 MB
    Throughput:       32.70 MB/sec
    Successful transactions: 1414
    Longest transaction:  10130.00 ms
    Shortest transaction:  160.00 ms
```

**Test 3**

`long_query_time = 2` (Log Only Queries > 2 sec)

`docker exec -it mysql-db mysql -uroot -prootpassword -e "SET GLOBAL long_query_time = 2;"`

Results:
```
    Data transferred:  1008.06 MB
    Throughput:       32.87 MB/sec
    Successful transactions: 1427
    Longest transaction:  8400.00 ms
    Shortest transaction:  200.00 ms
```

### Summary of results

| long_query_time    | Data transferred | Throughput | Successful transactions |
|--------------------|------------|------------|---------------------|
| **0s** | 892.92 MB  | 29.70 MB/sec  | 1264  |
| **0.3s**   | 998.88 MB  | 32.70 MB/sec  | 1414  |
| **2s**  | 1008.06 MB  | 32.87 MB/sec  | 1427  | 

- Lower `long_query_time` values log more queries, leading to higher system load.
- Higher `long_query_time` values reduce logging, improving throughput and response times.
