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

type tokenManager struct {
	client *fastly_ext.Client
}

// Manager returns an ability to AddTokens
func Manager(client *fastly_ext.Client) *tokenManager { // nolint
	return &tokenManager{
		client: client,
	}
}

// AddToken creates an API Token for a service(s) or an error
func (t *tokenManager) AddToken(req TokenRequest) (Token, error) {

	tokenInput := &fastly_ext.CreateTokenInput{
		Name:       req.Name,
		Password:   req.Password,
		TwoFAToken: req.TwoFAToken,
		Services:   req.Services,
		Scope:      req.Scope,
	}

	username, err := suppliedOrInteractive(req.Username, "Enter your Fastly username", terminal.GetInput())

	if err != nil {
		return Token{}, err
	}

	tokenInput.Username = username

	password, err := suppliedOrInteractive(req.Password, "Enter your Fastly password", terminal.GetInput())

	if err != nil {
		return Token{}, err
	}

	tokenInput.Password = password

	if req.RequireTwoFAToken {

		token, err := suppliedOrInteractive(req.TwoFAToken, "Enter your Fastly 2FA", terminal.GetInputSecret()) // nolint: govet

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
