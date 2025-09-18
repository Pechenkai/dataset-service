-- Users table
CREATE TABLE users
(
    id                BIGSERIAL PRIMARY KEY,
    username          VARCHAR(50)  NOT NULL,
    email             VARCHAR(255) NOT NULL,
    password          VARCHAR(255) NOT NULL,
    registration_date TIMESTAMPTZ  NOT NULL DEFAULT now(),
    country           VARCHAR(100) NOT NULL,
    is_blocked        BOOLEAN      NOT NULL DEFAULT FALSE,
    role              VARCHAR(10)  NOT NULL
);

ALTER TABLE users
    ADD CONSTRAINT uq_users_email UNIQUE (email),
    ADD CONSTRAINT chk_users_role CHECK (role IN ('guest', 'user', 'admin'));

-- Categories table
CREATE TABLE categories
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    description TEXT
);

ALTER TABLE categories
    ADD CONSTRAINT uq_categories_name UNIQUE (name);

-- Datasets table
CREATE TABLE datasets
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id    BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    is_public   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

ALTER TABLE datasets
    ADD CONSTRAINT fk_datasets_owner FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE RESTRICT,
    ADD CONSTRAINT fk_datasets_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT;

CREATE INDEX idx_datasets_owner ON datasets (owner_id);
CREATE INDEX idx_datasets_category_public_created ON datasets (category_id, is_public, created_at DESC);

-- Dataset versions table
CREATE TABLE dataset_versions
(
    id          BIGSERIAL PRIMARY KEY,
    number      VARCHAR(20) NOT NULL,
    upload_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    filepath    TEXT        NOT NULL,
    dataset_id  BIGINT      NOT NULL,
    change_log  TEXT
);

ALTER TABLE dataset_versions
    ADD CONSTRAINT fk_versions_dataset FOREIGN KEY (dataset_id) REFERENCES datasets (id) ON DELETE CASCADE;

CREATE INDEX idx_versions_dataset ON dataset_versions (dataset_id, upload_date DESC);

-- Metadata table
CREATE TABLE metadata
(
    id                 BIGSERIAL PRIMARY KEY,
    format             VARCHAR(50) NOT NULL,
    size               BIGINT      NOT NULL,
    tags               TEXT,
    dataset_version_id BIGINT      NOT NULL
);

ALTER TABLE metadata
    ADD CONSTRAINT chk_metadata_size CHECK (size > 0),
    ADD CONSTRAINT fk_metadata_version FOREIGN KEY (dataset_version_id) REFERENCES dataset_versions (id) ON DELETE CASCADE;

CREATE INDEX idx_metadata_version ON metadata (dataset_version_id);

-- Subscriptions table
CREATE TABLE subscriptions
(
    user_id    BIGINT NOT NULL,
    dataset_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, dataset_id)
);

ALTER TABLE subscriptions
    ADD CONSTRAINT fk_subscriptions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_subscriptions_dataset FOREIGN KEY (dataset_id) REFERENCES datasets (id) ON DELETE CASCADE;

CREATE INDEX idx_subscriptions_dataset ON subscriptions (dataset_id);

-- Reviews table
CREATE TABLE reviews
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT   NOT NULL,
    dataset_id BIGINT   NOT NULL,
    rating     SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    text       TEXT
);

ALTER TABLE reviews
    ADD CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_reviews_dataset FOREIGN KEY (dataset_id) REFERENCES datasets (id) ON DELETE CASCADE,
    ADD CONSTRAINT chk_reviews_rating CHECK (rating BETWEEN 1 AND 5);

CREATE INDEX idx_reviews_dataset ON reviews (dataset_id);
CREATE INDEX idx_reviews_user ON reviews (user_id);

-- Notifications table
CREATE TABLE notifications
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    dataset_id BIGINT NOT NULL,
    message    TEXT   NOT NULL,
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_notifications_dataset FOREIGN KEY (dataset_id) REFERENCES datasets (id) ON DELETE CASCADE;

CREATE INDEX idx_notifications_user ON notifications (user_id);