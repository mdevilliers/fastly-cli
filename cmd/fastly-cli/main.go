package main

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fastly-cli",
}

type config struct {
	FastlyAPIKey       string `envconfig:"FASTLY_API_KEY" default:""`
	FastlyUserName     string `envconfig:"FASTLY_USER_NAME" default:""`
	FastlyUserPassword string `envconfig:"FASTLY_USER_PASSWORD" default:""`
}

var globalConfig config

func initConfig() {

	err := envconfig.Process("", &globalConfig)

	if err != nil {
		panic(err)
	}
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyAPIKey, "fastly-api-key", globalConfig.FastlyAPIKey, "Fastly API Key")
	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyUserName, "fastly-user-name", globalConfig.FastlyUserName, "Fastly user name")
	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyUserPassword, "fastly-user-password", globalConfig.FastlyUserPassword, "Fastly user password")

	registerLaunchCommand(rootCmd)
	err := registerCreateCommand(rootCmd)

	if err != nil {
		log.Fatal(err)
	}

	err = registerTokenCommands(rootCmd)

	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
