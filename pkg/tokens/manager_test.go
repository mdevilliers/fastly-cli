package tokens

import (
	"testing"

	fastly_ext "github.com/mdevilliers/fastly-cli/pkg/fastly-ext"
	"github.com/stretchr/testify/require"
)

type mockTokenCreator struct {
	token *fastly_ext.Token
	err   error
	count int
}

func (m *mockTokenCreator) CreateToken(i *fastly_ext.CreateTokenInput) (*fastly_ext.Token, error) {
	m.count++
	return m.token, m.err
}

func Test_MissingInputsGathered(t *testing.T) {

	client := &mockTokenCreator{
		token: &fastly_ext.Token{},
	}

	count := 0
	tg := func(string) (string, error) {
		count++
		return "hello", nil
	}

	m := Manager(client, WithInput(tg), WithSecretInput(tg))

	// no username, password, twoFA
	_, err := m.AddToken(TokenRequest{
		RequireTwoFAToken: true,
	})

	require.Nil(t, err)
	require.Equal(t, 3, count)

	// no username, password
	count = 0
	_, err = m.AddToken(TokenRequest{
		RequireTwoFAToken: false,
	})

	require.Nil(t, err)
	require.Equal(t, 2, count)

	// no username
	count = 0
	_, err = m.AddToken(TokenRequest{
		RequireTwoFAToken: false,
		Username:          "John Smith",
	})

	require.Nil(t, err)
	require.Equal(t, 1, count)

	// all supplied
	count = 0
	_, err = m.AddToken(TokenRequest{
		RequireTwoFAToken: false,
		Username:          "John Smith",
		Password:          "secret",
	})

	require.Nil(t, err)
	require.Equal(t, 0, count)
}
