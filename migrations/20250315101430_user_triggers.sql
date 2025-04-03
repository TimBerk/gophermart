-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_user_balance()
    RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO balance (user_id) VALUES (NEW.id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_user_insert
    AFTER INSERT ON users
    FOR EACH ROW
EXECUTE FUNCTION create_user_balance();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS after_user_insert ON users;
DROP FUNCTION IF EXISTS create_user_balance;
-- +goose StatementEnd
