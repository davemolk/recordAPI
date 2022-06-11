CREATE INDEX IF NOT EXISTS albums_title_idx ON albums USING GIN (to_tsvector('simple', title)); 
CREATE INDEX IF NOT EXISTS albums_artist_idx ON albums USING GIN (to_tsvector('simple', artist)); 
CREATE INDEX IF NOT EXISTS albums_genres_idx ON albums USING GIN (genres);