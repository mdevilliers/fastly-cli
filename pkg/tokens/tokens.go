package tokens

import (
	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/terminal"
	"github.com/pkg/errors"
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

type tokenManager struct{}

// Manager returns an ability to AddTokens
func Manager() *tokenManager { // nolint
	return &tokenManager{}
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

	username, err := suppliedOrInteractive(req.Username, "Enter your Fastly username :", false)

	if err != nil {
		return Token{}, err
	}

	tokenInput.Username = username

	password, err := suppliedOrInteractive(req.Password, "Enter your Fastly password :", true)

	if err != nil {
		return Token{}, err
	}

	tokenInput.Password = password

	if req.RequireTwoFAToken {

		token, err := suppliedOrInteractive(req.TwoFAToken, "Enter your Fastly 2FA :", true) // nolint: govet

		if err != nil {
			return Token{}, err
		}

		tokenInput.TwoFAToken = token
	}

	resp, err := fastly_ext.CreateToken(tokenInput)

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

func suppliedOrInteractive(input, prompt string, secret bool) (string, error) {

	if input != "" {
		return input, nil
	}

	var value string
	var err error

	if secret {
		value, err = terminal.GetInputSecret(prompt)
	} else {
		value, err = terminal.GetInput(prompt)
	}

	if err != nil {
		return "", errors.Wrap(err, "error reading value")
	}

	return value, nil
}
