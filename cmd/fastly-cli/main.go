package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fastly-cli",
}

type config struct {
	FastlyAPIKey string `envconfig:"FASTLY_API_KEY" default:""`
}

var globalConfig config

func initConfig() {
	globalConfig = config{}
	err := envconfig.Process("", &globalConfig)

	if err != nil {
		panic(err)
	}
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&globalConfig.FastlyAPIKey, "fastlyAPIKey", "a", globalConfig.FastlyAPIKey, "FastlyAPIKey to use")

	registerLaunchCommand(rootCmd)
	err := registerCreateCommand(rootCmd)

	if err != nil {
		panic(err)
	}

	registerTokenCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
