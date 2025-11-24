-- Create messages table (for chat feature)
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for fast lookups
CREATE INDEX idx_messages_user_id ON messages(user_id);

-- Create index on created_at for ordering
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
