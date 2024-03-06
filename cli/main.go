package main

import (
	"github.com/bwiggs/spacetraders-go/cli/cmd"
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("ST")
	viper.AutomaticEnv()
}

func main() {
	cmd.Execute()
}
