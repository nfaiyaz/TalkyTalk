CREATE TABLE IF NOT EXISTS users (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(100) NOT NULL,
email VARCHAR(255) NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversations (
id BIGSERIAL PRIMARY KEY,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversation_participants (
conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
id BIGSERIAL PRIMARY KEY,
conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
body TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_participants_user_id
ON conversation_participants(user_id);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
ON messages(conversation_id, created_at);