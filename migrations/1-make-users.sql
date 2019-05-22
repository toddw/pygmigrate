CREATE TABLE users (
  id serial PRIMARY KEY,
  email VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  bio TEXT
);

CREATE INDEX users_email ON users (email);
