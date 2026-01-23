CREATE TABLE IF NOT EXISTS users (
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(64) UNIQUE,
	password_hash BYTEA,
	refresh_token TEXT UNIQUE,
	description VARCHAR(512)
);
