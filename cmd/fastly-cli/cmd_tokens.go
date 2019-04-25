package main

import (
	"fmt"

	"github.com/mdevilliers/fastly-cli/pkg/tokens"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
	"github.com/spf13/cobra"
)

func registerTokenCommands(root *cobra.Command) {

	tokenRoot := &cobra.Command{
		Use:   "tokens",
		Short: "Manage API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	listTokens := &cobra.Command{
		Use:   "all",
		Short: "List all API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			all, err := tokens.GetTokens(client, &tokens.GetTokensInput{})

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

	tokenRoot.AddCommand(listTokens)

	root.AddCommand(tokenRoot)
}
