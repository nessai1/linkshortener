CREATE TABLE IF NOT EXISTS hash_link (
    ID serial PRIMARY KEY,
    HASH varchar(255) NOT NULL UNIQUE,
    LINK varchar(255) NOT NULL
);