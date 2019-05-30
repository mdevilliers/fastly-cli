package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mdevilliers/fastly-cli/pkg/eavesdrop"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
	"github.com/spf13/cobra"
)

func registerEavesdropCommand(root *cobra.Command) error {

	var endpoint string
	var port int

	launchCommand := &cobra.Command{
		Use:   "eavesdrop",
		Short: "Listen in to your Fastly instance.",
		RunE: func(cmd *cobra.Command, args []string) error {

			client, err := fastly.NewClient(globalConfig.FastlyAPIKey)

			if err != nil {
				return errors.Wrap(err, "cannot create fastly client")
			}

			term := ""

			if len(args) > 0 {
				term = strings.Join(args, "")
			}

			service, err := getServiceWithPredicate(client, term)

			if err != nil {
				return err
			}

			if service == nil {
				fmt.Println("no service found")
				return nil
			}

			session := eavesdrop.NewSession(client, eavesdrop.SessionRequest{
				Endpoint:       endpoint,
				Port:           port,
				ServiceID:      service.ID,
				ServiceVersion: int(service.ActiveVersion),
			})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = session.StartListening()

			if err != nil {
				return err
			}

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt)

			<-stop

			if err := session.Dispose(ctx); err != nil {
				log.Print("error disposing session: ", err.Error())
				return err
			}

			return nil
		},
	}

	launchCommand.Flags().StringVar(&endpoint, "endpoint", endpoint, "endpoint to use for messages")
	launchCommand.Flags().IntVar(&port, "port", 443, "port to use for messages")

	err := launchCommand.MarkFlagRequired("endpoint")

	if err != nil {
		return err
	}

	root.AddCommand(launchCommand)
	return nil
}
