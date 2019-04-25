package tokens

import (
	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly"
	"github.com/mdevilliers/fastly-cli/pkg/terminal"
	"github.com/pkg/errors"
)

type tokenCreator interface {
	AddToken(TokenRequest) (Token, error)
}

type TokenRequest struct {
	Name              string
	Username          string
	Password          string
	Scope             string
	TwoFAToken        string
	RequireTwoFAToken bool
	Services          []string
}
type Token struct {
	Name        string
	ID          string
	Services    []string
	AccessToken string
}

type tokenManager struct{}

func Manager() *tokenManager {
	return &tokenManager{}
}

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

	password, err := suppliedOrInteractive(req.Username, "Enter your Fastly password :", true)

	if err != nil {
		return Token{}, err
	}

	tokenInput.Username = password

	if req.RequireTwoFAToken {

		token, err := suppliedOrInteractive(req.Username, "Enter your Fastly 2FA :", true)

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
