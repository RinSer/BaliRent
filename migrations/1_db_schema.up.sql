CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    phone TEXT NOT NULL,
    email TEXT NOT NULL,
    pwdhash TEXT NOT NULL,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS users_phone ON users(phone);
CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users(email);

CREATE TABLE IF NOT EXISTS properties (
    id SERIAL PRIMARY KEY,
    owner_id BIGINT REFERENCES users(id),
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    photos jsonb,
    description TEXT,
    price INTEGER NOT NULL,
    property_type SMALLINT NOT NULL,
    for_sell BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS properties_lat ON properties(latitude);
CREATE INDEX IF NOT EXISTS properties_long ON properties(longitude);

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    property_id BIGINT REFERENCES properties(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);