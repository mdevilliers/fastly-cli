package eavesdrop

import (
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
)

type serviceMutator func(version int) error

type builder struct {
	client  *fastly.Client
	service *fastly.Service
	version int
}

func Wrap(client *fastly.Client, service *fastly.Service) *builder {
	return &builder{
		client:  client,
		service: service,
	}
}

func (b *builder) clone() error {

	newVersion, err := b.client.CloneVersion(&fastly.CloneVersionInput{
		Service: b.service.ID,
		Version: int(b.service.ActiveVersion),
	})

	if err != nil {
		return errors.Wrap(err, "error cloning service")
	}

	b.version = newVersion.Number

	return nil
}

func (b *builder) Action(fn ...serviceMutator) error {

	err := b.clone()

	if err != nil {
		return err
	}

	for i := range fn {
		err = fn[i](b.version)
		if err != nil {
			return err
		}
	}

	return b.activate()

}

func (b *builder) activate() error {

	_, err := b.client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: b.service.ID,
		Version: b.version,
	})

	if err != nil {
		return errors.Wrap(err, "error activating version")
	}
	return nil

}
