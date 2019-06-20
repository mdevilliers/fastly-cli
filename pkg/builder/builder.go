package builder

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/pkg/errors"
)

// ServiceInfo contains the current fastly.Service ID and Version
type ServiceInfo struct {
	ID      string
	Version int
}

type serviceMutator func(current ServiceInfo) error

type builder struct {
	client         clonerActivator
	serviceID      string
	serviceVersion int
	latestVersion  int
}

type clonerActivator interface {
	CloneVersion(i *fastly.CloneVersionInput) (*fastly.Version, error)
	ActivateVersion(i *fastly.ActivateVersionInput) (*fastly.Version, error)
}

// New returns a builder instance that will `Clone`` the current version of a service,
// apply a series of changes and then `Activate` if no errors
func New(client clonerActivator, serviceID string, serviceVersion int) *builder {
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

// Apply takes a series of functions that mutate the current instance or return
// the first error
func (b *builder) Apply(fn ...serviceMutator) error {

	if err := b.clone(); err != nil {
		return err
	}

	info := ServiceInfo{ID: b.serviceID, Version: b.latestVersion}

	for i := range fn {
		if err := fn[i](info); err != nil {
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
