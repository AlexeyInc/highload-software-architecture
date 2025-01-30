
-- Table with some values for testing isolation levels

CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    value INT
);

INSERT INTO test_table (value) VALUES (10), (20), (30);