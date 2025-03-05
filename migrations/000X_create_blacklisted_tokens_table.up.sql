CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT blacklisted_tokens_token_unique UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS blacklisted_tokens_user_id_idx ON blacklisted_tokens (user_id);
CREATE INDEX IF NOT EXISTS blacklisted_tokens_expires_at_idx ON blacklisted_tokens (expires_at); 