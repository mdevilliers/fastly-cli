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
	client     *fastly.Client
	request    SessionRequest
	uniqueName string
}

type SessionRequest struct {
	Endpoint string
	Port     int
	Service  *fastly.Service
}

func NewSession(client *fastly.Client, request SessionRequest) *session { // nolint
	return &session{
		client:     client,
		request:    request,
		uniqueName: uniqueName(),
	}
}

func (s *session) Dispose(ctx context.Context) error {

	builder := Wrap(s.client, s.request.Service)
	return builder.Action(s.ensurePreviousSessionDoesNotExist)
}

func (s *session) StartListening() error {

	builder := Wrap(s.client, s.request.Service)

	createSyslog := func(newVersion int) error {

		_, err := s.client.CreateSyslog(&fastly.CreateSyslogInput{
			Service:     s.request.Service.ID,
			Version:     newVersion,
			Name:        s.uniqueName,
			Address:     s.request.Endpoint,
			Port:        uint(s.request.Port),
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
	listener, err := net.Listen("tcp", "localhost:8080")

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
		// read line by line from socket
		line, _, _ := reader.ReadLine()
		fmt.Println(string(line))
	}
}

func (s *session) ensurePreviousSessionDoesNotExist(version int) error {

	l, err := s.client.ListSyslogs(&fastly.ListSyslogsInput{
		Service: s.request.Service.ID,
		Version: version,
	})

	if err != nil {
		return errors.Wrap(err, "error listing syslogs")
	}

	for _, sys := range l {
		if sys.Name == s.uniqueName {

			err := s.client.DeleteSyslog(&fastly.DeleteSyslogInput{
				Service: s.request.Service.ID,
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
