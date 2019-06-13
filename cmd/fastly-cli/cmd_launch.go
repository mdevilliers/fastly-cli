package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/terminal"
	"github.com/pkg/errors"
	"github.com/sahilm/fuzzy"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

const (
	fastlyServiceURLPattern = "https://manage.fastly.com/configure/services/%s"
)

func registerLaunchCommand(root *cobra.Command) {

	launchCommand := &cobra.Command{
		Use:   "launch",
		Short: "Fuzzy search for a service and launch in browser.",
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

			if service != nil {
				url := fmt.Sprintf(fastlyServiceURLPattern, service.ID)
				err := open.Run(url)

				if err != nil {
					return err
				}

			}
			return nil
		},
	}

	root.AddCommand(launchCommand)
}

func getServiceWithPredicate(client *fastly.Client, term string) (*fastly.Service, error) {

	services, err := client.ListServices(&fastly.ListServicesInput{})

	if err != nil {
		return nil, errors.Wrap(err, "error searching fastly")
	}

	services = fuzzyMatch(services, term)

	if len(services) == 1 {
		return services[0], nil
	}

	// let the user choose which one to launch
	return terminal.NewServiceSelector()(services)

}

func fuzzyMatch(services []*fastly.Service, term string) []*fastly.Service {

	// nothing to filter by so return all sorted by name
	if term == "" {
		sort.Sort(byName(services))
		return services
	}

	results := fuzzy.FindFrom(term, servicesSource(services))

	// no fuzzy matches so return all sorted by name
	if len(results) == 0 {
		sort.Sort(byName(services))
		return services
	}

	ordered := []*fastly.Service{}
	for _, r := range results {
		ordered = append(ordered, services[r.Index])
	}

	return ordered
}

type byName []*fastly.Service

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type servicesSource []*fastly.Service

func (s servicesSource) String(i int) string { return s[i].Name }
func (s servicesSource) Len() int            { return len(s) }
