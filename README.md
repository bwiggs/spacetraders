# Database Migrations

Leverages `golang-migrate`. You need to also install with deps for sqlite3, which I don't think you can do via homebrew.

```console
go install -tags 'sqlite3 sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```