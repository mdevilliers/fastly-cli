package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/rivo/tview"
	"github.com/sahilm/fuzzy"
	"github.com/sethvargo/go-fastly/fastly"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

const (
	apiKeyEnvVar = "FASTLY_API_KEY"

	fastlyServiceURLPattern = "https://manage.fastly.com/configure/services/%s"
)

var launchCommand = &cobra.Command{
	Use:   "launch",
	Short: "Fuzzy search for a service and launch in browser.",
	RunE: func(cmd *cobra.Command, args []string) error {

		key := os.Getenv(apiKeyEnvVar)

		if key == "" {
			return errors.New("no api key found in environment.\n 'export FASTLY_API_KEY=xxxxx'")
		}

		client, err := fastly.NewClient(key)

		if err != nil {
			return errors.Wrap(err, "cannot create fastly client")
		}

		services, err := client.ListServices(&fastly.ListServicesInput{})

		if err != nil {
			return errors.Wrap(err, "error searching fastly")
		}

		term := ""

		if len(args) > 0 {
			term = strings.Join(args, "")
		}

		ordered := fuzzyMatch(services, term)

		// if there is only one url launch it
		if len(ordered) == 1 {
			return open.Run(fmt.Sprintf(fastlyServiceURLPattern, ordered[0].ID))
		}

		// let the user choose which one to launch
		app := tview.NewApplication()
		list := tview.NewList()
		for _, service := range ordered {

			url := fmt.Sprintf(fastlyServiceURLPattern, service.ID)
			list = list.AddItem(service.Name, fmt.Sprintf("Version: %v, Updated at : %v", service.ActiveVersion, service.UpdatedAt),
				0, // no binding
				func() {
					// nolint
					open.Run(url)
					app.Stop()
				})
		}

		list = list.AddItem("Quit", "", 'q', func() {
			app.Stop()
		})

		return app.SetRoot(list, true).Run()
	},
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
