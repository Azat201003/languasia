ALTER TABLE user_hobbies ADD UNIQUE (user_id, hobby_id);
ALTER TABLE user_languages ADD UNIQUE (user_id, language_id, is_known);
