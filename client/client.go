package client

import (
	"context"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/spf13/viper"
)

var client *api.Client

type TokenProvider struct{}

func (tp TokenProvider) AgentToken(ctx context.Context, operationName string) (api.AgentToken, error) {
	return api.AgentToken{Token: viper.GetString("API_TOKEN")}, nil
}

func Client() (*api.Client, error) {
	var err error

	if client == nil {
		viper.SetDefault("BASE_URL", "https://api.spacetraders.io/v2")
		client, err = api.NewClient(viper.GetString("BASE_URL"), TokenProvider{})
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
