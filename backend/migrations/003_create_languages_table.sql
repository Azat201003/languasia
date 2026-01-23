CREATE TABLE IF NOT EXISTS languages (
	language_id SERIAL PRIMARY KEY,
	name VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS user_languages (
	user_id SERIAL REFERENCES users(user_id),
	language_id SERIAL REFERENCES languages(language_id),
	level INTEGER,
	is_known BOOLEAN
);

