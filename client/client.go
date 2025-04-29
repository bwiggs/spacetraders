package client

import (
	"context"
	"net/http"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var theClient *Client

type TokenProvider struct{}

func (tp TokenProvider) AccountToken(ctx context.Context, operationName api.OperationName) (api.AccountToken, error) {
	return api.AccountToken{Token: viper.GetString("API_TOKEN")}, nil
}

func (tp TokenProvider) AgentToken(ctx context.Context, operationName api.OperationName) (api.AgentToken, error) {
	return api.AgentToken{Token: viper.GetString("API_TOKEN")}, nil
}

type Client struct {
	*api.Client
	rateLimiter *rate.Limiter
}

func GetClient() (api.Invoker, error) {
	if theClient == nil {

		// limit to 2 requests per second with a burst of 1
		// NOTE: use 2 requests per second ran a little too close to the sun and ended up returning some 429s.
		rateLimiter := rate.NewLimiter(rate.Every(600*time.Millisecond), 1)

		// use a rate-limited transport
		opt := api.WithClient(&http.Client{
			Timeout: 3 * time.Second,
			Transport: &RateLimitedTransport{
				Base:    http.DefaultTransport,
				Limiter: rateLimiter,
			},
		})

		viper.SetDefault("BASE_URL", "https://api.spacetraders.io/v2")
		apiClient, err := api.NewClient(viper.GetString("BASE_URL"), TokenProvider{}, opt)
		if err != nil {
			return nil, err
		}
		theClient = &Client{Client: apiClient, rateLimiter: rateLimiter}
	}

	return theClient, nil
}

func (c *Client) GetRateLimitPressure() float64 {
	return c.rateLimiter.Tokens()
}

// NavigateShip invokes navigate-ship operation.
//
// Navigate to a target destination. The ship must be in orbit to use this function. The destination
// waypoint must be within the same system as the ship's current location. Navigating will consume
// the necessary fuel from the ship's manifest based on the distance to the target waypoint.
// The returned response will detail the route information including the expected time of arrival.
// Most ship actions are unavailable until the ship has arrived at it's destination.
// To travel between systems, see the ship's Warp or Jump actions.
//
// POST /my/ships/{shipSymbol}/navigate
func (c *Client) NavigateShip(ctx context.Context, request api.OptNavigateShipReq, params api.NavigateShipParams) (*api.NavigateShipOK, error) {
	res, err := c.Client.NavigateShip(ctx, request, params)
	return res, err
}
