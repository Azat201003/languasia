CREATE TABLE IF NOT EXISTS chat_members (
	chat_id SERIAL REFERENCES chats(chat_id),
	user_id SERIAL REFERENCES users(user_id),
	is_admin BOOLEAN DEFAULT false
);

