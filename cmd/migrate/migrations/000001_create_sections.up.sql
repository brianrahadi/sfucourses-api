CREATE TABLE IF NOT EXISTS sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dept varchar(255) NOT NULL,
    number varchar(255) NOT NULL,
    title varchar(255) NOT NULL,
    term varchar(255) NOT NULL,
);