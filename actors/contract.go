package actors

import "github.com/bwiggs/spacetraders-go/api"

type Contract struct {
	*api.Contract
}

func (c *Contract) Revenue() int {
	return c.Terms.Payment.OnAccepted + c.Terms.Payment.OnAccepted
}
