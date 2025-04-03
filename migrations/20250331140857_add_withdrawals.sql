-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals(
     id SERIAL PRIMARY KEY,
     user_id BIGINT NOT NULL REFERENCES users(id),
     order_number BIGINT UNIQUE NOT NULL,
     sum DECIMAL,

     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
