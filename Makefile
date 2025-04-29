MIGRATE=migrate -source file://db/migrations -database sqlite3://db/spacetraders.db 

.PHONY: all run api-spec migrate db-up db-down db-drop db-reset ui

run:
	go run cli/main.go run

generate:
	go generate ./...

migrate: db-up

reset: db-delete db-up

db-delete:
	rm -f db/spacetraders.db

db-up:
	${MIGRATE} up

db-down:
	${MIGRATE} down 1

db-drop:
	${MIGRATE} drop

db-reset: db-drop db-up

ui:
	go run ui/*.go

ui-build:
	go build -o spacetraders-ui ./ui/*.go

wasm:
	go run github.com/hajimehoshi/wasmserve@latest ./ui
