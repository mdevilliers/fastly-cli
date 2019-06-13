package terminal

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/manifoldco/promptui"
)

type serviceSelector func(services []*fastly.Service) (*fastly.Service, error)

// NewServiceSelector returns a function that when executed with a
// slice of fastly.Service draws a widget to allow users to select
// a single fastly.Service
func NewServiceSelector() serviceSelector { // nolint
	return selectorWidget
}

func selectorWidget(services []*fastly.Service) (*fastly.Service, error) {

	all := indexable(services)

	prompt := promptui.Select{
		Label: "Select service",
		Items: all.Keys(),
	}

	_, result, err := prompt.Run()

	if err != nil {
		return nil, err
	}

	return all.ByKey(result), nil

}

// indexable wraps a slice of fastly.Services returning the
// set of keys and a lookup or an individual key
type indexable []*fastly.Service

func (s indexable) Keys() []string {

	r := []string{}

	for i := range s {
		r = append(r, s[i].Name)
	}
	return r
}

func (s indexable) ByKey(key string) *fastly.Service {

	for i := range s {
		if s[i].Name == key {
			return s[i]
		}
	}
	return nil
}
