package main

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/dictionary"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerSyncCommand(root *cobra.Command) error {

	//	var localFile, dict, service string

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Sync local dictionaries with Fastly.",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			syncer := dictionary.Manager(client)

			return syncer.Sync()
		},
	}

	/*launchCommand.Flags().StringVar(&externalEndpoint, "endpoint", externalEndpoint, "endpoint to use for messages from Fastly")
	launchCommand.Flags().IntVar(&externalPort, "port", 443, "port to use for messages from Fastly")

	launchCommand.Flags().StringVar(&localEndpoint, "local-endpoint", "localhost", "endpoint to use for messages from external endpoint")
	launchCommand.Flags().IntVar(&localPort, "local-port", 8080, "port to use for messages from external endpoint")

	err := launchCommand.MarkFlagRequired("endpoint")

	if err != nil {
		return err
	}
	*/
	root.AddCommand(syncCommand)
	return nil
}
