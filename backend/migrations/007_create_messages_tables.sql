CREATE TABLE IF NOT EXISTS messages (
	message_id SERIAL PRIMARY KEY,
	content VARCHAR(64),
	sender_id SERIAL REFERENCES users(user_id),
	chat_id SERIAL REFERENCES chats(chat_id)
);

