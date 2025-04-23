package actors

import (
	"time"

	"github.com/bwiggs/spacetraders-go/api"
)

type Contract struct {
	*api.Contract
}

func NewContract(c *api.Contract) *Contract {
	return &Contract{
		Contract: c,
	}
}

func (c *Contract) Revenue() int {
	return c.Terms.Payment.OnAccepted + c.Terms.Payment.OnAccepted
}

func (c *Contract) IsExpired() bool {
	return c.DeadlineToAccept.Value.Before(time.Now())
}
