CREATE TABLE IF NOT EXISTS messages (
	message_id SERIAL PRIMARY KEY,
	content VARCHAR(64),
	sender_id SERIAL REFERENCES users(user_id)
);

