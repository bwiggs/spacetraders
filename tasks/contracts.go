package tasks

import (
	"context"
	"math"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/pkg/errors"
)

func GetLatestContract(client api.Invoker) (*api.Contract, error) {
	page := 1
	limit := 20

	// get the latest contract from the api via the list endpoint
	res, err := client.GetContracts(context.TODO(), api.GetContractsParams{Page: api.NewOptInt(page), Limit: api.NewOptInt(limit)})
	if err != nil {
		return nil, errors.Wrap(err, "GetLatestContract: failed to get contracts")
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	// get the last page of contracts
	if res.Meta.Total > limit {
		page = int(math.Ceil(float64(res.Meta.Total) / float64(limit)))
		res, err = client.GetContracts(context.TODO(), api.GetContractsParams{Page: api.NewOptInt(page), Limit: api.NewOptInt(limit)})
		if err != nil {
			return nil, errors.Wrap(err, "GetLatestContract: failed to get contracts")
		}
	}
	return &res.Data[len(res.Data)-1], nil
}
