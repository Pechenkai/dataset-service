CREATE TABLE access_requests
(
    id         BIGSERIAL PRIMARY KEY,
    dataset_id BIGINT      NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     TEXT        NOT NULL CHECK (status IN ('pending', 'approved', 'denied')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (dataset_id, user_id)
);