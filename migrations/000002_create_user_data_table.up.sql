CREATE TYPE dtype AS ENUM ('TEXT', 'FILE', 'CARD', 'CREDENTIALS');

CREATE TABLE user_data (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data_type dtype NOT NULL,
    metadata JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    is_deleted BOOLEAN default False,
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_user_data_user_id ON user_data(user_id);
CREATE INDEX idx_user_data_data_type ON user_data(data_type);
CREATE INDEX idx_user_data_metadata_gin ON user_data USING GIN(metadata);