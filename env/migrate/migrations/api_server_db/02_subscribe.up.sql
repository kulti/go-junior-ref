BEGIN;

CREATE TABLE list_subscribers (
    list_id TEXT NOT NULL,
    email TEXT NOT NULL,
    processed_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (list_id) REFERENCES lists(id)
);

CREATE TABLE list_events (
    list_id TEXT NOT NULL,
    event_data JSONB NOT NULL,
    seq_num SERIAL NOT NULL,
    FOREIGN KEY (list_id) REFERENCES lists(id)
);

END;
