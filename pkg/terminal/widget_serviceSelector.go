package terminal

import (
	"github.com/manifoldco/promptui"
	"github.com/sethvargo/go-fastly/fastly"
)

type serviceSelector func(services []*fastly.Service) (*fastly.Service, error)

func NewServiceSelector() serviceSelector {
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
