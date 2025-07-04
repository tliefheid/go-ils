-- Books table
CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    isbn TEXT UNIQUE NOT NULL,
    publication_year INT NOT NULL,
    copies_total INT NOT NULL,
    copies_available INT NOT NULL
);
-- Members table
CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    contact TEXT
);
-- Borrowings table
CREATE TABLE IF NOT EXISTS borrowings (
    id SERIAL PRIMARY KEY,
    book_id INT REFERENCES books(id),
    member_id INT REFERENCES members(id),
    issue_date TIMESTAMP NOT NULL,
    return_date TIMESTAMP
);
