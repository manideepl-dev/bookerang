CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    username  VARCHAR(255) PRIMARY KEY,
    password  TEXT NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL
);

CREATE TABLE authors (
    author_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE books (
    book_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title     VARCHAR(255) NOT NULL,
    author_id UUID NOT NULL REFERENCES authors (author_id),
    CONSTRAINT books_title_author_key UNIQUE (title, author_id)
);

CREATE TABLE copies (
    copy_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id  UUID NOT NULL REFERENCES books (book_id),
    owner_id VARCHAR(255) NOT NULL REFERENCES users (username),
    CONSTRAINT copies_book_owner_key UNIQUE (book_id, owner_id)
);
