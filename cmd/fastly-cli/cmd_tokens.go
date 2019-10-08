package main

import (
	"fmt"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerTokenCommands(root *cobra.Command) error {

	tokenRoot := &cobra.Command{
		Use:   "tokens",
		Short: "Manage API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	var tokenName, service, tokenScope string

	addToken := &cobra.Command{
		Use:   "add",
		Short: "Add API token to existing service",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			if service == "" {
				return errors.New("supply a service name")
			}

			scope, err := stringToTokenScope(tokenScope)

			if err != nil {
				return err
			}

			tokenInput := tokens.TokenRequest{
				Name:     tokenName,
				Services: []string{service},
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
		},
	}

	addToken.Flags().StringVar(&service, "service-name", service, "name of service to create the API token for")
	addToken.Flags().StringVar(&tokenName, "token-name", tokenName, "name of the API token to create.")
	addToken.Flags().StringVar(&tokenScope, "token-scope", "global", "scope of the API token to create")

	err := markFlagsRequired(addToken, "service-name", "token-name")

	if err != nil {
		return err
	}

	listTokens := &cobra.Command{
		Use:   "all",
		Short: "List all API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			all, err := client.ListTokens()

			if err != nil {
				return errors.Wrap(err, "error listing tokens")
			}

			for _, t := range all {
				// TODO : use different output formats
				fmt.Println(t.Name, "[", t.ID, "]")
				fmt.Println("services :", t.Services)
			}

			return nil
		},
	}
	tokenRoot.AddCommand(addToken)
	tokenRoot.AddCommand(listTokens)

	root.AddCommand(tokenRoot)
	return nil
}

func stringToTokenScope(str string) (fastly.TokenScope, error) {

	potential := fastly.TokenScope(str)

	switch potential {
	case fastly.GlobalScope:
	case fastly.PurgeSelectScope:
	case fastly.PurgeAllScope:
	case fastly.GlobalReadScope:
		return potential, nil

	}
	return potential, fmt.Errorf("invalid TokenScope : %s", str)
}
