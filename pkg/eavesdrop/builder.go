package eavesdrop

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/pkg/errors"
)

type serviceInfo struct {
	ID      string
	Version int
}

type serviceMutator func(client *fastly.Client, current serviceInfo) error

type builder struct {
	client         *fastly.Client
	serviceID      string
	serviceVersion int
	latestVersion  int
}

// NewBuilder returns a builder instance that will `Clone`` the current version of a service,
// apply a series of changes and then `Activate` if no errors
func NewBuilder(client *fastly.Client, serviceID string, serviceVersion int) *builder {
	return &builder{
		client:         client,
		serviceID:      serviceID,
		serviceVersion: serviceVersion,
	}
}

func (b *builder) clone() error {

	newVersion, err := b.client.CloneVersion(&fastly.CloneVersionInput{
		Service: b.serviceID,
		Version: b.serviceVersion,
	})

	if err != nil {
		return errors.Wrap(err, "error cloning service")
	}

	b.latestVersion = newVersion.Number

	return nil
}

// Action takes a series of functions that mutate the current instance or return
// the first error
func (b *builder) Action(fn ...serviceMutator) error {

	err := b.clone()

	if err != nil {
		return err
	}

	for i := range fn {
		err = fn[i](b.client, serviceInfo{ID: b.serviceID, Version: b.latestVersion})
		if err != nil {
			return err
		}
	}

	return b.activate()

}

func (b *builder) activate() error {

	_, err := b.client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: b.serviceID,
		Version: b.latestVersion,
	})

	if err != nil {
		return errors.Wrap(err, "error activating version")
	}
	return nil
}
