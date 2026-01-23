CREATE TABLE IF NOT EXISTS hobbies (
	title VARCHAR(64),
	hobby_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS user_hobbies (
	user_id SERIAL REFERENCES users(user_id),
	hobby_id SERIAL REFERENCES hobbies(hobby_id)
);

