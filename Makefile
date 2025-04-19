MIGRATE=migrate -source file://db/migrations -database sqlite3://db/spacetraders.db 

.PHONY: all run api-spec migrate db-up db-down db-drop db-reset ui

run:
	go run *.go

api-spec:
	go generate ./...

migrate: migrate-up

db-up:
	${MIGRATE} up

db-down:
	${MIGRATE} down

db-drop:
	${MIGRATE} drop

db-reset: db-drop db-up

ui:
	go run ui/*.go

ui-build:
	go build -o ui/spacetraders-ui ./ui/*.go