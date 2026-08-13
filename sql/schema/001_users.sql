-- +goose Up
CREATE TABLE users(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    subscription_id TEXT
);
CREATE TABLE files(
    id INT PRIMARY KEY,
    path TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE TABLE files_users(
    file_id INT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY(file_id, user_id),
    FOREIGN KEY(file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE users;
DROP TABLE files;
DROP TABLE files_users;
