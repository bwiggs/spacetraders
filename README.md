# OpenAPI Spec

## Updating Local API

Pull the latest spec file

        wget https://raw.githubusercontent.com/SpaceTradersAPI/api-docs/refs/heads/main/reference/SpaceTraders.json -O SpaceTradersSpec.json

Regenerate the api client



# Database Migrations

## Requirements

Leverages `golang-migrate`. You need to also install with deps for sqlite3, which I don't think you can do via homebrew.

```console
go install -tags 'sqlite3 sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

then you can run `make db-up`

## Creating Migrations

```console
migrate create -dir db/migrations -ext sql new_migration_name
```

# Seeding System Data

Use the cli tool

This call might take a while since there are about 7000 systems and we can only fetch 20 at a time, and we're ratelimited at no more than 2 per second.

```
go run cli/main.go update systems
```