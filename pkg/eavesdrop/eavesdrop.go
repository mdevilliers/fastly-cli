package eavesdrop

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
)

type session struct {
	client     *fastly.Client
	request    SessionRequest
	created    bool // TODO : lock access
	uniqueName string
}

type SessionRequest struct {
	Endpoint       string
	Port           int
	ServiceID      string
	ServiceVersion int
}

func NewSession(client *fastly.Client, request SessionRequest) *session { // nolint
	return &session{
		client:     client,
		request:    request,
		uniqueName: "fastly-cli-delete-me", // TODO : make more unique
	}
}

func (s *session) Dispose(ctx context.Context) error {

	if s.created {

		err := s.client.DeleteSyslog(&fastly.DeleteSyslogInput{
			Service: s.request.ServiceID,
			Version: s.request.ServiceVersion,
			Name:    s.uniqueName,
		})

		if err != nil {
			return err
		}

	}

	return nil
}

func (s *session) StartListening() error {

	err := s.ensurePreviousSessionDoesNotExist()

	if err != nil {
		return err
	}

	created, err := s.client.CreateSyslog(&fastly.CreateSyslogInput{
		Service: s.request.ServiceID,
		Version: s.request.ServiceVersion,
		Name:    s.uniqueName,
		Address: s.request.Endpoint,
		Port:    uint(s.request.Port),
	})

	if err != nil {
		return errors.Wrap(err, "error creating syslog")
	}

	fmt.Println(created)
	s.created = true

	// start listening

	// stream to std out

	return nil
}

func (s *session) ensurePreviousSessionDoesNotExist() error {

	// idemnepotent syslog listener create or delete and create
	l, err := s.client.ListSyslogs(&fastly.ListSyslogsInput{
		Service: s.request.ServiceID,
		Version: s.request.ServiceVersion,
	})

	if err != nil {
		return errors.Wrap(err, "error listing syslogs")
	}

	for _, sys := range l {
		if sys.Name == s.uniqueName {
			err := s.client.DeleteSyslog(&fastly.DeleteSyslogInput{
				Service: s.request.ServiceID,
				Version: s.request.ServiceVersion,
				Name:    s.uniqueName,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}
