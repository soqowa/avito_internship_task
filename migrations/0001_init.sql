-- +goose Up
CREATE TABLE IF NOT EXISTS teams (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    team_id    TEXT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_id, is_active);

CREATE TABLE IF NOT EXISTS pull_requests (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    author_id  TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status     SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at  TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id      TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    slot       SMALLINT NOT NULL CHECK (slot IN (1, 2)),
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pr_id, slot),
    UNIQUE (pr_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user ON pr_reviewers(user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_pr ON pr_reviewers(user_id, pr_id);

-- +goose Down
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
