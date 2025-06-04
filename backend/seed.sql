-- Seed books
tRUNCATE books RESTART IDENTITY CASCADE;
INSERT INTO books (
        title,
        author,
        isbn,
        publication_year,
        copies_total,
        copies_available
    )
VALUES (
        'The Go Programming Language',
        'Alan A. A. Donovan',
        '9780134190440',
        2015,
        5,
        5
    ),
    (
        'Clean Code',
        'Robert C. Martin',
        '9780132350884',
        2008,
        3,
        3
    ),
    (
        'The Pragmatic Programmer',
        'Andrew Hunt',
        '9780201616224',
        1999,
        4,
        4
    ),
    (
        'To Kill a Mockingbird',
        'Harper Lee',
        '9780061120084',
        1960,
        2,
        2
    ),
    (
        '1984',
        'George Orwell',
        '9780451524935',
        1949,
        6,
        6
    );
-- Seed members
TRUNCATE members RESTART IDENTITY CASCADE;
INSERT INTO members (name, contact, member_id)
VALUES ('Alice Smith', 'alice@example.com', 'M001'),
    ('Bob Johnson', 'bob@example.com', 'M002'),
    ('Charlie Brown', 'charlie@example.com', 'M003'),
    ('Diana Prince', 'diana@example.com', 'M004');