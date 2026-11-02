CREATE TABLE stores (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    number      TEXT NOT NULL,
    sort_number TEXT NOT NULL,
    address     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT stores_number_not_empty  CHECK (length(btrim(number)) > 0),
    CONSTRAINT stores_address_not_empty CHECK (length(btrim(address)) > 0),
    CONSTRAINT stores_number_address_uniq UNIQUE (number, address)
);
CREATE INDEX stores_sort_idx ON stores (sort_number, id);

CREATE TABLE user_stores (
    user_id  BIGINT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, store_id)
);
CREATE INDEX user_stores_store_idx ON user_stores (store_id);

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE shifts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    store_id   BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT shifts_time_valid CHECK (starts_at < ends_at),
    CONSTRAINT shifts_max_24h    CHECK (ends_at - starts_at <= interval '24 hours'),

    CONSTRAINT shifts_no_overlap
        EXCLUDE USING gist (
            user_id WITH =,
            tstzrange(starts_at, ends_at, '[)') WITH &&
        )
);

CREATE INDEX shifts_store_window_idx ON shifts (store_id, starts_at);
CREATE INDEX shifts_user_window_idx  ON shifts (user_id, starts_at);