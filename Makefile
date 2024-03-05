MIGRATE=migrate -source file://db/migrations -database sqlite3://db/spacetraders.db 

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