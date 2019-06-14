package tokens

import (
	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly-ext"
	"github.com/mdevilliers/fastly-cli/pkg/terminal"
)

// TokenRequest contains data required to create a Fastly API token
type TokenRequest struct {
	Name              string
	Username          string
	Password          string
	Scope             string
	TwoFAToken        string
	RequireTwoFAToken bool
	Services          []string
}

// Token is a created Fastly API token
type Token struct {
	Name        string
	ID          string
	Services    []string
	AccessToken string
}

type tokenCreator interface {
	CreateToken(i *fastly_ext.CreateTokenInput) (*fastly_ext.Token, error)
}

type tokenManager struct {
	client   tokenCreator
	in       terminal.TextGatherer
	inSecret terminal.TextGatherer
}

type option func(*tokenManager)

// WithInput allows an override when asking for plain textutual input
func WithInput(in terminal.TextGatherer) option {
	return func(t *tokenManager) {
		t.in = in
	}
}

// WithSecretInput allows an override when asking for secrets
func WithSecretInput(in terminal.TextGatherer) option {
	return func(t *tokenManager) {
		t.inSecret = in
	}
}

// Manager returns an ability to AddTokens from the input supplied
// If required input is missing the manager will ask for the options via the terminal
func Manager(client tokenCreator, options ...option) *tokenManager { // nolint

	tm := tokenManager{
		client:   client,
		in:       terminal.GetInput(),
		inSecret: terminal.GetInputSecret(),
	}

	for _, o := range options {
		o(&tm)
	}
	return &tm
}

// AddToken creates an API Token for a service(s) or an error
func (t *tokenManager) AddToken(req TokenRequest) (Token, error) {

	tokenInput := &fastly_ext.CreateTokenInput{
		Username:   req.Username,
		Name:       req.Name,
		Password:   req.Password,
		TwoFAToken: req.TwoFAToken,
		Services:   req.Services,
		Scope:      req.Scope,
	}

	username, err := suppliedOrInteractive(req.Username, "Enter your Fastly username", t.in)

	if err != nil {
		return Token{}, err
	}

	tokenInput.Username = username

	password, err := suppliedOrInteractive(req.Password, "Enter your Fastly password", t.inSecret)

	if err != nil {
		return Token{}, err
	}

	tokenInput.Password = password

	if req.RequireTwoFAToken {

		token, err := suppliedOrInteractive(req.TwoFAToken, "Enter your Fastly 2FA", t.inSecret) // nolint: govet

		if err != nil {
			return Token{}, err
		}

		tokenInput.TwoFAToken = token
	}

	resp, err := t.client.CreateToken(tokenInput)

	if err != nil {
		return Token{}, err
	}

	return Token{
		Name:        resp.Name,
		ID:          resp.ID,
		Services:    resp.Services,
		AccessToken: resp.AccessToken,
	}, nil
}

func suppliedOrInteractive(initial, prompt string, g terminal.TextGatherer) (string, error) {

	if initial != "" {
		return initial, nil
	}

	return g(prompt)
}
