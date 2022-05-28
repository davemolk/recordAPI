CREATE TABLE IF NOT EXISTS albums (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(), 
    title text NOT NULL,
    artist text NOT NULL,
    genres text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);