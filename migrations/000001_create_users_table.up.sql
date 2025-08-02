CREATE TABLE IF NOT EXISTS users (
    username text PRIMARY KEY,
    date_of_birth timestamp(0) with time zone NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);