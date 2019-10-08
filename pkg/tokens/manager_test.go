package tokens

import (
	"testing"

	"github.com/fastly/go-fastly/fastly"
	"github.com/stretchr/testify/require"
)

type mockTokenCreator struct {
	token *fastly.Token
	err   error
	count int
}

func (m *mockTokenCreator) CreateToken(i *fastly.CreateTokenInput) (*fastly.Token, error) {
	m.count++
	return m.token, m.err
}

func Test_MissingInputsGathered(t *testing.T) {

	client := &mockTokenCreator{
		token: &fastly.Token{},
	}

	count := 0
	tg := func(string) (string, error) {
		count++
		return "hello", nil
	}

	m := Manager(client, WithInput(tg), WithSecretInput(tg))

	// no username, password, twoFA
	_, err := m.AddToken(TokenRequest{})

	require.Nil(t, err)
	require.Equal(t, 2, count)

	// no username, password
	count = 0
	_, err = m.AddToken(TokenRequest{})

	require.Nil(t, err)
	require.Equal(t, 2, count)

	// no username
	count = 0
	_, err = m.AddToken(TokenRequest{
		Username: "John Smith",
	})

	require.Nil(t, err)
	require.Equal(t, 1, count)

	// all supplied
	count = 0
	_, err = m.AddToken(TokenRequest{
		Username: "John Smith",
		Password: "secret",
	})

	require.Nil(t, err)
	require.Equal(t, 0, count)
}
