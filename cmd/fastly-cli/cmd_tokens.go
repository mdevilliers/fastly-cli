package main

import (
	"fmt"

	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly-ext"
	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
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
	var enable2FA bool

	addToken := &cobra.Command{
		Use:   "add",
		Short: "Add API token to existing service",
		RunE: func(cmd *cobra.Command, args []string) error {

			if service == "" {
				return errors.New("supply a service name")
			}

			tokenInput := tokens.TokenRequest{
				Name:              tokenName,
				Services:          []string{service},
				Scope:             tokenScope,
				RequireTwoFAToken: enable2FA,
				Username:          globalConfig.FastlyUserName,
				Password:          globalConfig.FastlyUserPassword,
			}

			tokenManager := tokens.Manager()
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
	addToken.Flags().BoolVar(&enable2FA, "enable-2FA", true, "use 2FA. If enabled you will be asked to provide a token when creating an API user")

	err := addToken.MarkFlagRequired("service-name")

	if err != nil {
		return err
	}

	err = addToken.MarkFlagRequired("token-name")

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

			all, err := fastly_ext.GetTokens(client, &fastly_ext.GetTokensInput{})

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
