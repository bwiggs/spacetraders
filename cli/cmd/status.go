package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwiggs/spacetraders-go/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "returns system/server information",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.GetClient()
		if err != nil {
			log.Fatal(err)
		}

		st, err := client.GetStatus(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Version:", st.Version)
		fmt.Println("Status:", st.Status)
		// fmt.Println("Leaderboards:", st.Leaderboards)
		fmt.Println("Stats:", st.Stats)
		resetDate, err := time.Parse(time.RFC3339, st.ServerResets.Next)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Next Reset:", fmt.Sprintf("%s (%s):", st.ServerResets.Next, time.Until(resetDate)))
		fmt.Println()
		for _, a := range st.Announcements {
			fmt.Println(a.Title, "-", a.Body)
		}

	},
}
