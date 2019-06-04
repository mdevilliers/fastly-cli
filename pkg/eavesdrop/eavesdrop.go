package eavesdrop

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/user"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
)

type session struct {
	sessionOptions
	client     *fastly.Client
	uniqueName string
}

type option func(*sessionOptions)

// WithExternalBindng ads the facility to override the external TCP service
func WithExternalBinding(endpoint string, port int) option { // nolint
	return func(r *sessionOptions) {
		r.ExternalEndpoint = endpoint
		r.ExternalPort = port
	}
}

// WithLocalBindng ads the facility to override the local binding to the TCP service
func WithLocalBinding(endpoint string, port int) option { // nolint
	return func(r *sessionOptions) {
		r.LocalEndpoint = endpoint
		r.LocalPort = port
	}
}

// sessionRequest contains the information for creating a session
type sessionOptions struct {
	ExternalEndpoint string
	ExternalPort     int
	LocalEndpoint    string
	LocalPort        int
	Service          *fastly.Service
}

// NewSession returns a connction to an existing service
func NewSession(client *fastly.Client, service *fastly.Service, options ...option) *session { // nolint

	defaultSessionOptions := &sessionOptions{
		LocalEndpoint: "localhost",
		LocalPort:     8080,
		Service:       service,
	}

	for _, o := range options {
		o(defaultSessionOptions)
	}

	return &session{
		sessionOptions: *defaultSessionOptions,
		client:         client,
		uniqueName:     uniqueName(),
	}
}

func (s *session) Dispose(ctx context.Context) error {
	builder := Wrap(s.client, s.Service)
	return builder.Action(s.ensurePreviousSessionDoesNotExist)
}

func (s *session) StartListening() error {

	builder := Wrap(s.client, s.Service)

	createSyslog := func(newVersion int) error {

		_, err := s.client.CreateSyslog(&fastly.CreateSyslogInput{
			Service:     s.Service.ID,
			Version:     newVersion,
			Name:        s.uniqueName,
			Address:     s.ExternalEndpoint,
			Port:        uint(s.ExternalPort),
			MessageType: "blank",
			Format:      `{ "type": "req","service_id": "%{req.service_id}V","request_id": "%{req.http.fastly-soc-x-request-id}V","start_time": "%{time.start.sec}V","fastly_info": "%{fastly_info.state}V", "datacenter": "%{server.datacenter}V","client_ip": "%a", "req_method": "%m", "req_uri": "%{cstr_escape(req.url)}V", "req_h_host": "%{cstr_escape(req.http.Host)}V", "req_h_referer": "%{cstr_escape(req.http.referer)}V", "req_h_user_agent": "%{cstr_escape(req.http.User-Agent)}V", "req_h_accept_encoding": "%{cstr_escape(req.http.Accept-Encoding)}V", "req_header_bytes": "%{req.header_bytes_read}V", "req_body_bytes": "%{req.body_bytes_read}V", "resp_status": "%{resp.status}V", "resp_bytes": "%{resp.bytes_written}V", "resp_header_bytes": "%{resp.header_bytes_written}V", "resp_body_bytes": "%{resp.body_bytes_written}V" }`,
		})

		if err != nil {
			return errors.Wrap(err, "error creating syslog")
		}
		return nil
	}

	err := builder.Action(s.ensurePreviousSessionDoesNotExist, createSyslog)

	if err != nil {
		return err
	}

	// start listening
	binding := fmt.Sprintf("%s:%d", s.LocalEndpoint, s.LocalPort)
	listener, err := net.Listen("tcp", binding)

	if err != nil {
		return errors.Wrap(err, "error creating listener")
	}

	defer listener.Close() // nolint:errcheck

	connection, err := listener.Accept()

	if err != nil {
		return errors.Wrap(err, "error accepting connection")
	}
	go handleConnection(connection)
	return nil
}

func handleConnection(connection net.Conn) {
	defer connection.Close() // nolint: errcheck
	reader := bufio.NewReader(connection)

	for {
		line, _, _ := reader.ReadLine()
		fmt.Println(string(line))
	}
}

func (s *session) ensurePreviousSessionDoesNotExist(version int) error {

	l, err := s.client.ListSyslogs(&fastly.ListSyslogsInput{
		Service: s.Service.ID,
		Version: version,
	})

	if err != nil {
		return errors.Wrap(err, "error listing syslogs")
	}

	for _, sys := range l {
		if sys.Name == s.uniqueName {

			err := s.client.DeleteSyslog(&fastly.DeleteSyslogInput{
				Service: s.Service.ID,
				Version: version,
				Name:    s.uniqueName,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func uniqueName() string {

	user, err := user.Current()

	if err != nil {
		return "fastly-cli-unknown-user"
	}

	return fmt.Sprintf("fastly-cli-%s", user.Username)

}
