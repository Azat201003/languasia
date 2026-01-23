CREATE INDEX idx_users_username ON users USING gin(username);
CREATE INDEX idx_user_hobbies_user_id ON user_hobbies(user_id);
CREATE INDEX idx_user_languages_user_id ON user_languages(user_id);
CREATE INDEX idx_user_languages_is_known ON user_languages(is_known);
