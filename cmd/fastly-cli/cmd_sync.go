package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/dictionary"
	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly-ext"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerSyncCommand(root *cobra.Command) error {

	var localFile, filetype, dict, service string

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Sync local CSV files with Fastly edge dictionaries.",
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

			services, err := client.ListServices(&fastly.ListServicesInput{})

			if err != nil {
				return errors.Wrap(err, "error searching fastly for services")
			}

			version := 0
			serviceID := ""
			for _, s := range services {
				if s.Name == service {
					version = int(s.ActiveVersion)
					serviceID = s.ID
				}
			}

			if version == 0 {
				return fmt.Errorf("cannot find service : %s", service)
			}

			dictInstance, err := client.GetDictionary(&fastly.GetDictionaryInput{
				Service: serviceID,
				Version: version,
				Name:    dict,
			})

			if err != nil {
				return errors.Wrap(err, "error getting dictionary ID")
			}

			extendedClient := fastly_ext.NewExtendedClient(client)
			syncer := dictionary.Manager(extendedClient, dictionary.WithLocalReader(reader), dictionary.WithRemoteDictionary(serviceID, dictInstance.ID))

			return syncer.Sync()
		},
	}

	syncCommand.Flags().StringVar(&localFile, "path", localFile, "path to file")
	syncCommand.Flags().StringVar(&filetype, "file-type", "CSV", "type of file")
	syncCommand.Flags().StringVar(&dict, "dict", dict, "name of dictionary to update")
	syncCommand.Flags().StringVar(&service, "service", service, "name of service to update")

	err := markFlagsRequired(syncCommand, "path", "dict", "service")

	if err != nil {
		return err
	}

	root.AddCommand(syncCommand)
	return nil
}
