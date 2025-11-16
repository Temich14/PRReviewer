CREATE TABLE IF NOT EXISTS teams
(
    id VARCHAR(36) PRIMARY KEY,
    team_name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS users
(
    id VARCHAR(36) PRIMARY KEY,
    username TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS team_members
(
    team_id VARCHAR(36),
    user_id VARCHAR(36),
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pull_requests
(
    id VARCHAR(36) PRIMARY KEY,
    pr_name TEXT,
    author_id VARCHAR (36),
    status TEXT,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers
(
    pr_id VARCHAR(36),
    reviewer_id VARCHAR(36),
    PRIMARY KEY (pr_id, reviewer_id),
    FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE
);