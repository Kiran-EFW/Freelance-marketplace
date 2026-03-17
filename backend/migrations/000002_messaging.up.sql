-- Conversations between users
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID REFERENCES jobs(id),
    participant_1 UUID NOT NULL REFERENCES users(id),
    participant_2 UUID NOT NULL REFERENCES users(id),
    last_message_at TIMESTAMPTZ,
    last_message_preview TEXT,
    is_archived_1 BOOLEAN NOT NULL DEFAULT false,
    is_archived_2 BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_p1 ON conversations(participant_1);
CREATE INDEX idx_conversations_p2 ON conversations(participant_2);
CREATE INDEX idx_conversations_job ON conversations(job_id);
CREATE INDEX idx_conversations_last_msg ON conversations(last_message_at DESC);
CREATE UNIQUE INDEX idx_conversations_participants ON conversations(
    LEAST(participant_1, participant_2),
    GREATEST(participant_1, participant_2),
    COALESCE(job_id, '00000000-0000-0000-0000-000000000000')
);

-- Messages within conversations
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'text', -- text, image, quote, system
    attachment_url TEXT,
    attachment_type VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_unread ON messages(conversation_id, is_read) WHERE is_read = false;

CREATE TRIGGER set_updated_at BEFORE UPDATE ON conversations FOR EACH ROW EXECUTE FUNCTION update_updated_at();
