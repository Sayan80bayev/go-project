CREATE TABLE IF NOT EXISTS subscriptions (
                                             id UUID PRIMARY KEY,
                                             follower_id UUID NOT NULL,
                                             followee_id UUID NOT NULL,
                                             approved BOOLEAN NOT NULL DEFAULT TRUE,
                                             created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                             deleted_at TIMESTAMP WITH TIME ZONE,
                                             UNIQUE (follower_id, followee_id)
    );

CREATE INDEX IF NOT EXISTS i_followee ON subscriptions (followee_id, deleted_at);
CREATE INDEX IF NOT EXISTS i_follower ON subscriptions (follower_id, deleted_at);