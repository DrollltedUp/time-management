CREATE TABLE IF NOT EXISTS stores (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    number      TEXT NOT NULL,
    sort_number TEXT NOT NULL,
    address     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT stores_number_not_empty  CHECK (length(btrim(number)) > 0),
    CONSTRAINT stores_address_not_empty CHECK (length(btrim(address)) > 0),
    CONSTRAINT stores_number_address_uniq UNIQUE (number, address)
);
CREATE INDEX IF NOT EXISTS stores_sort_idx ON stores (sort_number, id);

CREATE TABLE IF NOT EXISTS manager_stores (
    user_id  BIGINT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, store_id)
);
CREATE INDEX IF NOT EXISTS manager_stores_store_idx ON manager_stores (store_id);