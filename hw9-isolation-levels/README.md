
### Percona
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ✅  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ✅  |
| **Serializable**     | ✅  | ❌  | ❌  | ❌  |

### Postgres
| Isolation Level    | Lost Update | Dirty Read | Non-Repeatable Read | Phantom Read |
|--------------------|------------|------------|---------------------|--------------|
| **Read Uncommitted** | ✅  | ✅  | ✅  | ✅  |
| **Read Committed**   | ✅  | ❌  | ✅  | ✅  |
| **Repeatable Read**  | ✅  | ❌  | ❌  | ✅  |
| **Serializable**     | ✅  | ❌  | ❌  | ❌  |


---
 

## Lost Update

### Step-by-Step Breakdown of the Simulation**

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

The expectation was that at the SERIALIZABLE isolation level, Transaction B would wait for Transaction A to commit before proceeding. However, both database engines allowed concurrent reads of the same record, resulting in Transaction A overwriting Transaction B’s update, leading to a Lost Update.

To avoid a Lost Update, using `FOR UPDATE` alone is sufficient.


## Dirty Read




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
docker exec -it ac3839ba675e2c17c7f70c5fe1dc13cfb4aa2434b6d9457182dd947e96e705f8 mysql -u root -p
```
```
SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE; SHOW TRANSACTION ISOLATION LEVEL;
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