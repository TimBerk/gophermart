DB_URI = "postgres://gophermart:gophermart@localhost:5434/gophermart?sslmode=disable"

DB_DRIVER=postgres
DB_STRING="user=gophermart dbname=gophermart sslmode=disable password=gophermart host=localhost port=5434"
MIGRATIONS_DIR=./migrations

lint:
	go mod verify
	go vet ./...
	staticcheck ./...
build:
	go build -o cmd/gophermart/gophermart cmd/gophermart/main.go
	chmod +x cmd/gophermart/gophermart
test:
	go test -count=1 -cover ./...
run:
	go run cmd/gophermart/main.go -d $(DB_URI)
runa:
	./cmd/accrual/accrual_windows_amd64 -a ":8081" -d $(DB_URI)

# Work with DB container
dbu:
	docker-compose up -d
dbd:
	docker-compose -f docker-compose.yml down
	docker volume rm gophermart_db

# Apply all migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) up

# Rollback the latest migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) down

# Rollback all migrations
migrate-down-all:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) reset

# Check the status of migrations
migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) status

# Create a new migration file
migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

add-autotest:
	git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git

update-autotest:
	git fetch template && git checkout template/master .github
