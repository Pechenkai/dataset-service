-- Users table
CREATE TABLE users
(
    id                BIGSERIAL PRIMARY KEY,
    username          VARCHAR(50)  NOT NULL,
    email             VARCHAR(255) NOT NULL UNIQUE,
    password          VARCHAR(255) NOT NULL,
    registration_date TIMESTAMPTZ  NOT NULL DEFAULT now(),
    country           VARCHAR(100) NOT NULL,
    is_blocked        BOOLEAN      NOT NULL DEFAULT FALSE,
    role              VARCHAR(10)  NOT NULL CHECK (role IN ('guest', 'user', 'admin'))
);

-- Categories table
CREATE TABLE categories
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Datasets table
CREATE TABLE datasets
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id    BIGINT       NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    category_id BIGINT       NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    is_public   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
-- Index for common listing
CREATE INDEX idx_datasets_owner ON datasets (owner_id);
CREATE INDEX idx_datasets_category_public_created ON datasets (category_id, is_public, created_at DESC);

-- Dataset versions table
CREATE TABLE dataset_versions
(
    id          BIGSERIAL PRIMARY KEY,
    number      VARCHAR(20) NOT NULL,
    upload_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    filepath    TEXT        NOT NULL,
    dataset_id  BIGINT      NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    change_log  TEXT
);
CREATE INDEX idx_versions_dataset ON dataset_versions (dataset_id, upload_date DESC);

-- Metadata table
CREATE TABLE metadata
(
    id                 BIGSERIAL PRIMARY KEY,
    format             VARCHAR(50) NOT NULL,
    size               BIGINT      NOT NULL CHECK (size > 0),
    tags               TEXT,
    dataset_version_id BIGINT      NOT NULL REFERENCES dataset_versions (id) ON DELETE CASCADE
);
CREATE INDEX idx_metadata_version ON metadata (dataset_version_id);

-- Subscriptions table
CREATE TABLE subscriptions
(
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    dataset_id BIGINT      NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, dataset_id)
);
CREATE INDEX idx_subscriptions_dataset ON subscriptions (dataset_id);

-- Reviews table
CREATE TABLE reviews
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    dataset_id BIGINT      NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    rating     SMALLINT    NOT NULL CHECK (rating BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    text       TEXT
);
CREATE INDEX idx_reviews_dataset ON reviews (dataset_id);
CREATE INDEX idx_reviews_user ON reviews (user_id);

-- Notifications table
CREATE TABLE notifications
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    dataset_id BIGINT      NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    message    TEXT        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_user ON notifications (user_id);