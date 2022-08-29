.PHONY: clean swag build

APP_NAME = shoppinglist-backend
BUILD_DIR = $(PWD)/build
MIGRATIONS_FOLDER = $(PWD)/db/migrations

clean:
	rm -rf $(BUILD_DIR)

build: clean swag
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/api/main.go

swag:
	swag init --output api --generalInfo cmd/api/main.go
