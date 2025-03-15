CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    published_date DATE
);

INSERT INTO books (title, author, published_date) VALUES
('The Pragmatic Programmer', 'Andrew Hunt', '1999-10-20'),
('Clean Code', 'Robert C. Martin', '2008-08-11'),
('Designing Data-Intensive Applications', 'Martin Kleppmann', '2017-03-16'),
('The Psychedelic Experience', 'Timothy Leary', '1964-01-01'),
('Turn On, Tune In, Drop Out', 'Timothy Leary', '1967-01-01'),
('Your Brain is God', 'Timothy Leary', '1988-01-01'),
('PIHKAL: A Chemical Love Story', 'Alexander Shulgin', '1991-01-01'),
('TIHKAL: The Continuation', 'Alexander Shulgin', '1997-01-01'),
('Meditations', 'Marcus Aurelius', '180-01-01');