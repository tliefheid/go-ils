-- Seed books
TRUNCATE books RESTART IDENTITY CASCADE;
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
        1,
        1
    ),
    (
        'Clean Code',
        'Robert C. Martin',
        '9780132350884',
        2008,
        1,
        1
    ),
    (
        'The Pragmatic Programmer',
        'Andrew Hunt',
        '9780201616224',
        1999,
        1,
        1
    ),
    (
        'To Kill a Mockingbird',
        'Harper Lee',
        '9780061120084',
        1960,
        1,
        1
    ),
    (
        '1984',
        'George Orwell',
        '9780451524935',
        1949,
        1,
        1
    );
-- Seed members
TRUNCATE members RESTART IDENTITY CASCADE;
INSERT INTO members (name, contact)
VALUES (
        'Alice Smith',
        'alice@example.com'
    ),
    (
        'Bob Johnson',
        'bob @example.com'
    ),
    (
        'Charlie Brown',
        'charlie @example.com '
    ),
    (
        'Diana Prince',
        'diana @example.com'
    );
