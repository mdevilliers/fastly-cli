package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/eavesdrop"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerEavesdropCommand(root *cobra.Command) error {

	var externalEndpoint, localEndpoint string
	var externalPort, localPort int

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

			session := eavesdrop.NewSession(client,
				service,
				eavesdrop.WithExternalBinding(externalEndpoint, externalPort),
				eavesdrop.WithLocalBinding(localEndpoint, localPort),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = session.StartListening()

			if err != nil {
				return err
			}

			stop := make(chan os.Signal, 2)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

			<-stop

			if err := session.Dispose(ctx); err != nil {
				log.Print("error disposing session: ", err.Error())
				return err
			}

			return nil
		},
	}

	launchCommand.Flags().StringVar(&externalEndpoint, "endpoint", externalEndpoint, "endpoint to use for messages from Fastly")
	launchCommand.Flags().IntVar(&externalPort, "port", 443, "port to use for messages from Fastly")

	launchCommand.Flags().StringVar(&localEndpoint, "local-endpoint", "localhost", "endpoint to use for messages from external endpoint")
	launchCommand.Flags().IntVar(&localPort, "local-port", 8080, "port to use for messages from external endpoint")

	err := launchCommand.MarkFlagRequired("endpoint")

	if err != nil {
		return err
	}

	root.AddCommand(launchCommand)
	return nil
}
