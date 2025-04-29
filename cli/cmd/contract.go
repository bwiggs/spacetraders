package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/bwiggs/spacetraders-go/client"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	rootCmd.AddCommand(contractCmd)
}

var contractCmd = &cobra.Command{
	Use:   "contracts",
	Short: "returns contracts",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(args) == 1 {
			if args[0] == "new" {
				err = listContracts()
			}
		} else {
			err = listContracts()
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}

func newContract() error {
	return nil
}

func listContracts() error {
	client, err := client.GetClient()
	if err != nil {
		return err
	}

	res, err := client.GetContracts(context.TODO(), api.GetContractsParams{})
	if err != nil {
		return err
	}

	spew.Dump(res.Data)

	p := message.NewPrinter(language.English)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "FACTION\tACCEPTED\tSTATUS\tON ACCEPTED\tON FULFILLED")
	for _, c := range res.Data {
		fmt.Fprintf(w, "%s\t%t\t%t\t%s\t%s\n", c.FactionSymbol, c.Accepted, c.Fulfilled, p.Sprintf("%d", c.Terms.Payment.OnAccepted), p.Sprintf("%d", c.Terms.Payment.OnFulfilled))
	}
	w.Flush()

	return nil
}
