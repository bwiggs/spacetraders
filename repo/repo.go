package repo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Repo struct {
	db *sql.DB
}

var repo *Repo

func GetRepo() (*Repo, error) {
	if repo == nil {
		var err error
		repo, err = NewRepo(viper.GetString("DB"))
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func NewRepo(dbPath string) (*Repo, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &Repo{
		db: db,
	}, nil
}

func (repo *Repo) Close() {
	repo.db.Close()
}
