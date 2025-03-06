#!/bin/bash

# Define MySQL credentials
MYSQL_ROOT_PWD="111"
DB_NAME="mydb"

# Trap SIGTERM and exit gracefully
trap "echo 'Stopping script...'; exit 0" SIGTERM

# Initialize table only once
docker exec mysql_master sh -c "export MYSQL_PWD=$MYSQL_ROOT_PWD; mysql -u root $DB_NAME -e 'DROP TABLE IF EXISTS code; CREATE TABLE code(code INT);'"

echo "Starting periodic inserts and verification..."
counter=1

while true; do
    # Insert new data into the master
    docker exec mysql_master sh -c "export MYSQL_PWD=$MYSQL_ROOT_PWD; mysql -u root $DB_NAME -e 'INSERT INTO code VALUES ($counter);'"

    echo "Inserted $counter into mysql_master.code"

    # Wait a bit for replication
    sleep 1

    # Check replication on both slaves
    SLAVE1_COUNT=$(docker exec mysql_slave sh -c "export MYSQL_PWD=$MYSQL_ROOT_PWD; mysql -u root $DB_NAME -e 'SELECT COUNT(*) FROM code;' 2>/dev/null | tail -n 1")
    SLAVE2_COUNT=$(docker exec mysql_slave2 sh -c "export MYSQL_PWD=$MYSQL_ROOT_PWD; mysql -u root $DB_NAME -e 'SELECT COUNT(*) FROM code;' 2>/dev/null | tail -n 1")

    echo "Slave1 count: $SLAVE1_COUNT | Slave2 count: $SLAVE2_COUNT"

    # Compare counts
    if [[ "$SLAVE1_COUNT" != "$counter" || "$SLAVE2_COUNT" != "$counter" ]]; then
        echo "⚠️ Replication issue detected! Master inserted $counter, but slaves have ($SLAVE1_COUNT, $SLAVE2_COUNT)"
    fi

    # Wait before next insert
    sleep 2
    ((counter++))
done