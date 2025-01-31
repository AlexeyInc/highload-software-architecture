## Project Overview

This project demonstrates how different transaction isolation levels affect parallel database access in Percona (MySQL/InnoDB) and PostgreSQL. It explores key concurrency issues such as Lost Updates, Dirty Reads, Non-Repeatable Reads, and Phantom Reads by running parallel transactions with different isolation levels.

App built using Go and MySQL env is set up with docker-compose.

### How to Use

1. Run `docker-compose up -d` to start the project.
2. Execute `go run ./scr/cmd/main.go` from the root directory to launch the application.

### Percona
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ✅  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ❌  |
| **Serializable**     | ❌ - Deadlock  | ❌  | ❌  | ❌  |

### Postgres
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ❌  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ❌ - transaction error  | ❌  | ❌  | ❌  |
| **Serializable**     | ❌ - transaction error  | ❌  | ❌  | ❌  |


---
 
## Lost Update

####  CURLs to simulate lost update with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"                           

curl -X POST "http://localhost:8080/lost-update?dbDriver={{percona or postgres}}&isolation=READ%20UNCOMMITTED" \
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
curl -X POST "http://localhost:8080/lost-update?dbDriver={{percona or postgres}}&isolation=READ%20COMMITTED" \ 
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
curl -X POST "http://localhost:8080/lost-update?dbDriver={{percona or postgres}}&isolation=REPEATABLE%20READ" \
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
curl -X POST "http://localhost:8080/lost-update?dbDriver={{percona or postgres}}&isolation=SERIALIZABLE" \     
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
```

#### Results

**Percona**
 img

**Postgres**
 img


#### Summary
 
In **Percona**, attempting to simulate a Lost Update at the SERIALIZABLE isolation level results in a deadlock (occur because of row-level locks.).
In **Postgres**, the same test produces a “could not serialize access due to concurrent update” error at both REPEATABLE READ and SERIALIZABLE levels (because of Multi-Version Concurrency Control (MVCC) conflict detection).

To avoid a Lost Update, using `FOR UPDATE` alone is sufficient.

---

## Dirty Read

####  CURLs to simulate non-repeatable read with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"                           

curl "http://localhost:8080/dirty-read?dbDriver={{percona or postgres}}&newValue=300&isolation=READ%20UNCOMMITTED"
curl "http://localhost:8080/dirty-read?dbDriver={{percona or postgres}}&newValue=200&isolation=READ%20COMMITTED"
curl "http://localhost:8080/dirty-read?dbDriver={{percona or postgres}}&newValue=200&isolation=REPEATABLE%20READ"
curl "http://localhost:8080/dirty-read?dbDriver={{percona or postgres}}&newValue=200&isolation=SERIALIZABLE" 
```

#### Results

**Percona**
 img

**Postgres**
 img


#### Summary

Both databases prevent non-repeatable reads at REPEATABLE READ and SERIALIZABLE isolation levels. PostgreSQL relies on MVCC snapshots, resulting in a different SERIALIZABLE behavior in logs compared to Percona, which uses row-level locking, effectively preventing concurrency anomalies by forcing transactions to execute sequentially.

Expectation is that Transaction B should be able to read Transaction A’s uncommitted changes on the isolation level `READ UNCOMMITTED`.

**Percona** allow dirty reads when the isolation level is set to `READ UNCOMMITTED`. 
**PostgreSQL** `READ UNCOMMITTED` is internally treated as `READ COMMITTED`, which prevents dirty reads.

---

## Non-Repeatable read

####  CURLs to simulate non-repeatable read with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"                           

curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=1&isolation=READ%20UNCOMMITTED"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=2&isolation=READ%20COMMITTED"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=3&isolation=REPEATABLE%20READ"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=4&isolation=SERIALIZABLE"
```

#### Results

**Percona**

![Screenshot 2025-01-31 at 12 01 08](https://github.com/user-attachments/assets/48e8b83b-64ad-4f85-92c7-5cf9517866ff)


**Postgres**

![Screenshot 2025-01-31 at 12 05 09](https://github.com/user-attachments/assets/700c8f4b-2761-47f1-b1af-52a73b80761f)


#### Summary

Both databases prevent non-repeatable reads at REPEATABLE READ and SERIALIZABLE isolation levels. PostgreSQL relies on MVCC snapshots, resulting in a different SERIALIZABLE behavior in logs compared to Percona, which uses row-level locking, effectively preventing concurrency anomalies by forcing transactions to execute sequentially.


## Phantom Read

####  CURLs to simulate phantom-read with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"       

curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=READ%20UNCOMMITTED"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=READ%20COMMITTED"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=REPEATABLE%20READ"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=SERIALIZABLE"
```

#### Results

**Percona**

![Screenshot 2025-01-31 at 10 21 30](https://github.com/user-attachments/assets/34b77aef-0e29-4a20-8919-8e90a8224169)


**Postgres**

![Screenshot 2025-01-31 at 11 17 41](https://github.com/user-attachments/assets/12473da7-e436-4a4d-807d-fd91532f0fc7)


#### Summary

Both databases prevent phantom reads at REPEATABLE READ and SERIALIZABLE isolation levels. PostgreSQL relies on MVCC snapshots, resulting in a different SERIALIZABLE behavior in logs compared to Percona, which uses row-level locking, effectively preventing concurrency anomalies by forcing transactions to execute sequentially.

---

### Queries to access database and change isolation levels manually

**Percona:**

```
docker exec -it <container_id> mysql -u root -p
```
```
SET SESSION TRANSACTION ISOLATION LEVEL SERIALIZABLE; 
SELECT @@transaction_isolation;
SHOW VARIABLES LIKE 'autocommit';
SET autocommit = 0;
```

**Postgres:**

```
docker exec -it <container_id> psql -U testuser -d testdb
```
```
SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE; SHOW TRANSACTION ISOLATION LEVEL;
```
