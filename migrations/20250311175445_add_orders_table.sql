-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED',
    'UNDEFINED'
);

CREATE TABLE IF NOT EXISTS orders(
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number BIGINT UNIQUE NOT NULL,
    status order_status NOT NULL DEFAULT 'NEW',
    accrual DECIMAL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
-- +goose StatementEnd
