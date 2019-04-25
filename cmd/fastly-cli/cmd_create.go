package main

import (
	"fmt"

	"github.com/mdevilliers/fastly-cli/pkg/terminal"
	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
	"github.com/spf13/cobra"
)

func registerCreateCommand(root *cobra.Command) error {

	var serviceName string
	var tokenName string
	var createAPIKey bool
	var enable2FA bool

	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a new Fastly service",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			service, err := client.CreateService(&fastly.CreateServiceInput{
				Name: serviceName,
			})

			if err != nil {
				return errors.Wrap(err, "error creating Service")
			}

			// TODO : output to different formats
			fmt.Println("service created")
			fmt.Println("service ID : ", service.ID)

			if !createAPIKey {
				return nil
			}

			if tokenName == "" {
				tokenName = serviceName
			}

			tokenInput := &tokens.CreateTokenInput{
				Name:     tokenName,
				Services: []string{service.ID},
				Scope:    "global",
			}

			username, err := terminal.GetInput("Enter your Fastly username :")

			if err != nil {
				return errors.Wrap(err, "error reading username")
			}

			tokenInput.Username = username

			password, err := terminal.GetInputSecret("Enter your Fastly password :")

			if err != nil {
				return errors.Wrap(err, "error reading password")
			}

			tokenInput.Password = password

			if enable2FA {

				token, err := terminal.GetInputSecret("Enter your 2FA Token :") // nolint: govet

				if err != nil {
					return errors.Wrap(err, "error reading 2FA Token")
				}

				tokenInput.TwoFAToken = token

			}

			token, err := tokens.CreateToken(tokenInput)

			if err != nil {
				return errors.Wrap(err, "error creating token")
			}

			fmt.Println("API :", token.Name)
			fmt.Println("API access token", token.AccessToken)

			return nil
		}}

	createCommand.Flags().StringVar(&serviceName, "service-name", serviceName, "name of service to create")
	createCommand.Flags().StringVar(&tokenName, "token-name", tokenName, "name of the API token to create. Defaults to the service-name if not supplied")
	createCommand.Flags().BoolVar(&createAPIKey, "create-api-token", true, "create an API token")
	createCommand.Flags().BoolVar(&enable2FA, "enable-2FA", true, "use 2FA. If enabled you will be asked to provide a token when creating an API user")

	err := createCommand.MarkFlagRequired("service-name")

	if err != nil {
		return err
	}

	root.AddCommand(createCommand)

	return nil
}
