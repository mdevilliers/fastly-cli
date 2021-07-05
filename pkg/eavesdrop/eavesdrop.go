package eavesdrop

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os/user"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/builder"
	"github.com/pkg/errors"
)

type session struct {
	sessionOptions
	client *fastly.Client
}

type option func(*sessionOptions)

// WithExternalBinding adds the facility to override the external TCP service
func WithExternalBinding(endpoint string, port int) option { // nolint
	return func(r *sessionOptions) {
		r.ExternalEndpoint = endpoint
		r.ExternalPort = port
	}
}

// WithLocalBinding adds the facility to override the local binding to the TCP service
func WithLocalBinding(endpoint string, port int) option { // nolint
	return func(r *sessionOptions) {
		r.LocalEndpoint = endpoint
		r.LocalPort = port
	}
}

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
	}
}

func (s *session) Dispose(ctx context.Context) error {

	// get latest service
	latest, err := s.client.GetServiceDetails(&fastly.GetServiceInput{
		ID: s.Service.ID,
	})

	if err != nil {
		return errors.Wrap(err, "error getting latest service")
	}
	instance := builder.New(s.client, latest.ID, latest.ActiveVersion.Number)
	return instance.Apply(s.ensurePreviousSessionDoesNotExist)
}

func (s *session) StartListening() error {

	instance := builder.New(s.client, s.Service.ID, int(s.Service.ActiveVersion))

	createSyslog := func(current builder.ServiceInfo) error {

		_, err := s.client.CreateSyslog(&fastly.CreateSyslogInput{
			Service:     current.ID,
			Version:     current.Version,
			Name:        uniqueName(),
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

	err := instance.Apply(s.ensurePreviousSessionDoesNotExist, createSyslog)

	if err != nil {
		return err
	}

	// start TCP server
	binding := fmt.Sprintf("%s:%d", s.LocalEndpoint, s.LocalPort)
	listener, err := net.Listen("tcp", binding)

	if err != nil {
		return errors.Wrap(err, "error creating listener")
	}

	go listen(listener)

	return nil
}

func listen(listener net.Listener) {

	defer listener.Close() // nolint:errcheck
	for {

		fmt.Println("waiting for messages...")
		connection, err := listener.Accept()

		if err != nil {
			break
		}
		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {

	defer connection.Close() // nolint: errcheck
	reader := bufio.NewReader(connection)

	for {
		line, _, err := reader.ReadLine()

		if err != nil {
			fmt.Println(err.Error())
			if errors.Is(err, io.EOF) {
				break
			}
		}

		fmt.Println(string(line))
	}
}

func (s *session) ensurePreviousSessionDoesNotExist(current builder.ServiceInfo) error {

	syslogName := uniqueName()
	l, err := s.client.ListSyslogs(&fastly.ListSyslogsInput{
		Service: current.ID,
		Version: current.Version,
	})

	if err != nil {
		return errors.Wrap(err, "error listing syslogs")
	}

	for _, sys := range l {
		if sys.Name == syslogName {

			return s.client.DeleteSyslog(&fastly.DeleteSyslogInput{
				Service: current.ID,
				Version: current.Version,
				Name:    syslogName,
			})
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
