CREATE TABLE IF NOT EXISTS likes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (user_id, post_id)
);

CREATE INDEX IF NOT EXISTS i_user_id ON likes (user_id, deleted_at);
CREATE INDEX IF NOT EXISTS i_post_id ON likes (post_id, deleted_at);
