#!/bin/bash

# Database connection details
CONTAINER_NAME="postgresql-b"
DB_USER="postgres"
DB_NAME="books_db"

docker exec -it $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    published_date DATE
);
"

# Generate 10,000 random records and insert them
echo "Generating 10,000 random records..."

SQL_FILE="random_books.sql"

# Generate SQL file
cat > $SQL_FILE <<EOF
INSERT INTO books (title, author, published_date) VALUES
EOF

for i in {1..10000}
do
    TITLE="Book $RANDOM"
    AUTHOR="Author $RANDOM"
    DATE="$(shuf -i 1500-2025 -n 1)-$(shuf -i 1-12 -n 1)-$(shuf -i 1-28 -n 1)"

    if [ $i -eq 10000 ]; then
        echo "('$TITLE', '$AUTHOR', '$DATE');" >> $SQL_FILE
    else
        echo "('$TITLE', '$AUTHOR', '$DATE')," >> $SQL_FILE
    fi
done

# Execute the SQL file inside PostgreSQL
docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < $SQL_FILE

# Cleanup
rm $SQL_FILE

echo "Insertion complete!"