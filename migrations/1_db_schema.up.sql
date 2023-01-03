CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    phone TEXT NOT NULL,
    pwdhash TEXT NOT NULL,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE rent_or_sell AS ENUM ('rent', 'sell');
CREATE TYPE property_type AS ENUM ('room', 'flat', 'house');

CREATE TABLE IF NOT EXISTS properties (
    id SERIAL PRIMARY KEY,
    owner_id BIGINT REFERENCES users(id),
    address TEXT NOT NULL,
    photos jsonb,
    description TEXT,
    price INTEGER NOT NULL,
    property_type property_type NOT NULL,
    rent_or_sell rent_or_sell NOT NULL,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    property_id BIGINT REFERENCES properties(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    when_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);