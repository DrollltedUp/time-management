CREATE TABLE users(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOTNULL UNIQUE CHECK (email = lower(email)),
    password_hasg TEXT NOT NULL,
    full_name TEXT NOT NULL CHECK (char_length(full_name) BETWEEN 1 AND 100),
    role TEXT NOT NULL DEFAULT 'employee' CHECK (role IN ('employee', 'manager')),
    active BOOLEAN NOT NULL DEFAULT true
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);