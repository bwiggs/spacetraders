package client

import (
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimitedTransport struct {
	Base    http.RoundTripper
	Limiter *rate.Limiter
}

func (r *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := r.Limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	return r.Base.RoundTrip(req)
}
