package main

import (
	"bufio"
	"encoding/csv"
	"os"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/dictionary"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerSyncCommand(root *cobra.Command) error {

	var localFile, filetype, dict, service string

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Sync local dictionaries with Fastly edge dictionaries.",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			// we only know about csv files
			csvFile, err := os.Open(localFile) // nolint : gosec 'localFile' path is passed in via the user

			if err != nil {
				return errors.Wrap(err, "error opening csv file")
			}

			reader := csv.NewReader(bufio.NewReader(csvFile))

			syncer := dictionary.Manager(client, dictionary.WithLocalReader(reader), dictionary.WithRemoteDictionary(service, dict))

			return syncer.Sync()
		},
	}

	syncCommand.Flags().StringVar(&localFile, "path", localFile, "path to file")
	syncCommand.Flags().StringVar(&filetype, "file-type", "CSV", "type of file")
	syncCommand.Flags().StringVar(&dict, "dict", dict, "dictionary to update")
	syncCommand.Flags().StringVar(&service, "service", service, "service to update")

	err := markFlagsRequired(syncCommand, "path", "dict", "service")

	if err != nil {
		return err
	}

	root.AddCommand(syncCommand)
	return nil
}
