CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE users
    ADD COLUMN location GEOGRAPHY(POINT, 4326);

CREATE INDEX users_location_idx ON users USING GIST (location);
