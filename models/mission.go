package models

import (
	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/repo"
)

type TradeMission struct {
	client *api.Client
	repo   *repo.Repo
}

// func NewTradeMission(client *api.Client, repo *repo.Repo) *TradeMission {
// 	return &Mission{}
// }
