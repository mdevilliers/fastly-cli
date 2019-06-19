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

	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyAPIKey, "fastly-api-key", globalConfig.FastlyAPIKey, "Fastly API Key (export FASTLY_API_KEY=xxxx)")
	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyUserName, "fastly-user-name", globalConfig.FastlyUserName, "Fastly user name (export FASTLY_USER_NAME=xxxx)")
	rootCmd.PersistentFlags().StringVar(&globalConfig.FastlyUserPassword, "fastly-user-password", globalConfig.FastlyUserPassword, "Fastly user password (export FASTLY_USER_PASSWORD=xxxx)")

	err := registerChildCommands(rootCmd,
		registerEavesdropCommand,
		registerSyncCommand,
		registerCreateCommand,
		registerTokenCommands,
		registerLaunchCommand)

	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

type attachChildCommand func(*cobra.Command) error

func registerChildCommands(root *cobra.Command, childern ...attachChildCommand) error {

	for _, cmd := range childern {
		err := cmd(root)

		if err != nil {
			return err
		}
	}
	return nil
}

func markFlagsRequired(cmd *cobra.Command, flags ...string) error {

	for _, f := range flags {
		err := cmd.MarkFlagRequired(f)

		if err != nil {
			return err
		}

	}
	return nil
}
