-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS balance(
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current DECIMAL DEFAULT 0,
    withdrawn DECIMAL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance;
-- +goose StatementEnd
