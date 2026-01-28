CREATE TYPE chat_type AS ENUM ('Direct');

CREATE TABLE IF NOT EXISTS chats (
	chat_id SERIAL PRIMARY KEY,
	title VARCHAR(32),
	type chat_type DEFAULT 'Direct'
);

