CREATE TABLE posts (
  id serial PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  post TEXT
)
