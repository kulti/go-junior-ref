BEGIN;

CREATE TABLE lists (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE items (
    id TEXT NOT NULL PRIMARY KEY,
    list_id TEXT NOT NULL,
    name TEXT NOT NULL,
    done BOOLEAN DEFAULT false,
    FOREIGN KEY (list_id) REFERENCES lists(id)
);

COMMIT;
