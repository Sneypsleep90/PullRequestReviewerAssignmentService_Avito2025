-- +goose Up
-- +goose StatementBegin
CREATE TABLE pull_request_reviewers (
    pr_id VARCHAR(50) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL REFERENCES users(id),
    PRIMARY KEY (pr_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pull_request_reviewers;
-- +goose StatementEnd
