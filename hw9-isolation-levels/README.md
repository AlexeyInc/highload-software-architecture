
### Percona
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ✅  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ❌  |
| **Serializable**     | Deadlock  | ❌  | ❌  | ❌  |

### Postgres
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ❌  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ❌  |
| **Serializable**     | ✅  | ❌  | ❌  | ❌  |


---
 
## Lost Update

### Step-by-Step breakdown of simulation**

**Transaction A**

- Starts a transaction.
- Reads the value from `test_table WHERE id = 1`.
- Sleeps for 2 seconds (simulating a delay in processing).
- Adds `newValue` to the retrieved value.
- Updates the value in `test_table`.
- Commits the transaction.

**Transaction B**

- Starts a transaction 1 second after Transaction A (simulating concurrent access).
- Reads the value from `test_table WHERE id = 1` before Transaction A commits.
- Computes the new value based on the stale read.
- Updates the database overwriting changes made by Transaction A.
- Commits the transaction.

### Summary

The expectation was that at the SERIALIZABLE isolation level, Transaction B would wait for Transaction A to commit before proceeding. However, both database engines allowed concurrent reads of the same record, resulting in Transaction A overwriting Transaction B’s update, leading to a Lost Update.

To avoid a Lost Update, using `FOR UPDATE` alone is sufficient.

---

## Dirty Read

### Step-by-Step breakdown of simulation**

- Transaction A (Writer) updates the value (e.g. to 300) without committing.
- Transaction B (Reader) tries to read the value before A commits.
Expected Result: Transaction B should see 300.
- Transaction A rollbacks changes

### Summary

Expectation is that Transaction B should be able to read Transaction A’s uncommitted changes on the isolation level `READ UNCOMMITTED`.

*InnoDB* (the default storage engine in MySQL and **Percona**) does not actually allow dirty reads even when the isolation level is set to `READ UNCOMMITTED`.
(To get dirty reads we can switch to *MyISAM*, which does not support transactions but allows reading uncommitted data).

Same applies for **PostgreSQL**. `READ UNCOMMITTED` is internally treated as READ COMMITTED, which prevents dirty reads.

---

## Non-Repeatable read

###  CURLs to simulate non-repeatable read with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"                           

curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=1&isolation=READ%20UNCOMMITTED"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=2&isolation=READ%20COMMITTED"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=3&isolation=REPEATABLE%20READ"
curl "http://localhost:8080/non-repeatable-read?dbDriver={{percona or postgres}}&newValue=4&isolation=SERIALIZABLE"
```

### Results

**Percona**

![Screenshot 2025-01-31 at 12 01 08](https://github.com/user-attachments/assets/48e8b83b-64ad-4f85-92c7-5cf9517866ff)


**Postgres**

![Screenshot 2025-01-31 at 12 05 09](https://github.com/user-attachments/assets/700c8f4b-2761-47f1-b1af-52a73b80761f)


### Summary

Both databases prevent non-repeatable reads at REPEATABLE READ and SERIALIZABLE isolation levels. PostgreSQL relies on MVCC snapshots, resulting in a different SERIALIZABLE behavior in logs compared to Percona, which uses row-level locking, effectively preventing concurrency anomalies by forcing transactions to execute sequentially.

---

## Phantom Read

###  CURLs to simulate phantom-read with different isoalation levels:

```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"       

curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=READ%20UNCOMMITTED"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=READ%20COMMITTED"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=REPEATABLE%20READ"
curl "http://localhost:8080/phantom-read?dbDriver={{percona or postgres}}&isolation=SERIALIZABLE"
```

### Results

**Percona**

![Screenshot 2025-01-31 at 10 21 30](https://github.com/user-attachments/assets/34b77aef-0e29-4a20-8919-8e90a8224169)


**Postgres**

![Screenshot 2025-01-31 at 11 17 41](https://github.com/user-attachments/assets/12473da7-e436-4a4d-807d-fd91532f0fc7)


### Summary

Both databases prevent phantom reads at REPEATABLE READ and SERIALIZABLE isolation levels. PostgreSQL relies on MVCC snapshots, resulting in a different SERIALIZABLE behavior in logs compared to Percona, which uses row-level locking, effectively preventing concurrency anomalies by forcing transactions to execute sequentially.

### Queries to access database and change isolation levels

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
