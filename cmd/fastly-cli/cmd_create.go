package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
	"github.com/spf13/cobra"
)

func registerCreateCommand(root *cobra.Command) error {

	var serviceName string
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

			// TODO : output to differenct formats
			fmt.Println("service created")
			fmt.Println("service ID : ", service.ID)

			if !createAPIKey {
				return nil
			}

			tokenInput := &tokens.CreateTokenInput{
				Name:     fmt.Sprintf("API-%s", serviceName),
				Services: []string{service.ID},
				Scope:    "global",
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Println("username:")
			input, err := reader.ReadString('\n')
			if err != nil {
				return errors.Wrap(err, "error reading username")
			}
			tokenInput.Username = strings.Replace(input, "\n", "", -1)

			fmt.Println("password:")
			input, err = reader.ReadString('\n')
			if err != nil {
				return errors.Wrap(err, "error reading password")
			}
			tokenInput.Password = strings.Replace(input, "\n", "", -1)

			if enable2FA {
				fmt.Println("2FA:")
				input, err = reader.ReadString('\n')
				if err != nil {
					return errors.Wrap(err, "error reading 2FA")
				}
				tokenInput.TwoFAToken = strings.Replace(input, "\n", "", -1)

			}

			fmt.Println("tokenInput", tokenInput)

			token, err := tokens.CreateToken(tokenInput)

			if err != nil {
				return errors.Wrap(err, "error creating token")
			}

			fmt.Println(token.Name)
			fmt.Println(token.AccessToken)

			return nil
		}}

	createCommand.Flags().StringVar(&serviceName, "service-name", serviceName, "Name of service to create")
	createCommand.Flags().BoolVar(&createAPIKey, "create-api-user", true, "Create an API user")
	createCommand.Flags().BoolVar(&enable2FA, "enable-2FA", true, "Use 2FA. If enabled you will be asked to provide a token when creating an API user")

	err := createCommand.MarkFlagRequired("service-name")

	if err != nil {
		return err
	}

	root.AddCommand(createCommand)

	return nil
}
