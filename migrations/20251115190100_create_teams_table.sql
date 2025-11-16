-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    name VARCHAR(50) PRIMARY KEY
);

-- Добавление внешнего ключа для связи пользователей с командами
ALTER TABLE users ADD CONSTRAINT fk_user_team 
FOREIGN KEY (team_name) REFERENCES teams(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_user_team;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
