MAIN_PATH = ./main.go

TEST_LIST_PATH := ./list
TEST_TODO_PATH := ./todo
TEST_MIDDLEWARE_PATH := ./api

export DB_USER=postgres
export DB_PWD=example
export DB_NAME=postgres
export DB_PORT=5433
export DB_HOST=localhost

test-list:
	echo "Running list unit tests"
	go test -v $(TEST_LIST_PATH)

test-todo:
	echo "Running todo unit tests"
	go test -v $(TEST_TODO_PATH)

test-middleware:
	echo "Running middleware unit tests"
	go test -v $(TEST_MIDDLEWARE_PATH)

build:
	echo "Building application"
	go build -o ./bin/main.go $(MAIN_PATH)

debug:
	go run -gcflags "all=-N -l" $(MAIN_PATH)

debug2:
	dlv debug $(MAIN_PATH)

run:
	echo "Running application ..."
	@go run $(MAIN_PATH)
