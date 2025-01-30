
### Percona
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ❌  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ✅  |
| **Serializable**     | ✅  | ❌  | ❌  | ❌  |

### Postgres
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ❌  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ✅  |
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

## Non-Repeatable Read

### Queries to access database and change isolation levels

**Percona:**

```
docker exec -it <container_id> mysql -u root -p
```
```
SET SESSION TRANSACTION ISOLATION LEVEL SERIALIZABLE; SELECT @@transaction_isolation;
```

**Postgres:**

```
docker exec -it <container_id> psql -U testuser -d testdb
```
```
SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE; SHOW TRANSACTION ISOLATION LEVEL;
```

###  CURL to init db tables:
```
curl "http://localhost:8080/init-db?dbDriver={{percona or postgres}}"  
```

###  CURLs to simulate two concurrent transactions for lost update:

**Percona:**

```
curl -X POST "http://localhost:8080/lost-update?dbDriver=percona" \
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
```

**Postgres:**

```
curl -X POST "http://localhost:8080/lost-update?dbDriver=postgres" \
     -H "Content-Type: application/json" \
     -d '{"valueA": 200, "valueB": 500}'
```