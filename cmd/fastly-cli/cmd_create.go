package main

import (
	"fmt"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerCreateCommand(root *cobra.Command) error {

	var serviceName string
	var tokenName string
	var tokenScope string
	var createAPIKey bool

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

			scope, err := stringToTokenScope(tokenScope)

			if err != nil {
				return err
			}

			tokenInput := tokens.TokenRequest{
				Name:     tokenName,
				Services: []string{service.ID},
				Scope:    scope,
				Username: globalConfig.FastlyUserName,
				Password: globalConfig.FastlyUserPassword,
			}

			tokenManager := tokens.Manager(client)
			token, err := tokenManager.AddToken(tokenInput)

			if err != nil {
				return errors.Wrap(err, "error creating token")
			}

			fmt.Println("API :", token.Name)
			fmt.Println("API access token", token.AccessToken)

			return nil
		}}

	createCommand.Flags().StringVar(&serviceName, "service-name", serviceName, "name of service to create")
	createCommand.Flags().StringVar(&tokenName, "token-name", tokenName, "name of the API token to create. Defaults to the service-name if not supplied")
	createCommand.Flags().StringVar(&tokenScope, "token-scope", "global", "scope of the API token to create")

	createCommand.Flags().BoolVar(&createAPIKey, "create-api-token", false, "create an API token")

	err := createCommand.MarkFlagRequired("service-name")

	if err != nil {
		return err
	}

	root.AddCommand(createCommand)

	return nil
}
