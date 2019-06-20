package builder

import (
	"errors"
	"testing"

	"github.com/fastly/go-fastly/fastly"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	cloneVersionCalled    bool
	activateVersionCalled bool

	cloneVersioner    func(i *fastly.CloneVersionInput) (*fastly.Version, error)
	activateVersioner func(i *fastly.ActivateVersionInput) (*fastly.Version, error)
}

func (m *mockClient) CloneVersion(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	m.cloneVersionCalled = true
	return m.cloneVersioner(i)
}

func (m *mockClient) ActivateVersion(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	m.activateVersionCalled = true
	return m.activateVersioner(i)
}

func Test_CloneAndActivateCalled_Success(t *testing.T) {

	client := &mockClient{
		cloneVersioner: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 1}, nil
		},
		activateVersioner: func(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 2}, nil
		},
	}

	builder := New(client, "foo", 1)
	err := builder.Apply() // doing notthing shouldn't error

	require.Nil(t, err)
	require.True(t, client.cloneVersionCalled)
	require.True(t, client.activateVersionCalled)

}

func Test_CloneFailsWithError(t *testing.T) {

	client := &mockClient{
		cloneVersioner: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
			return nil, errors.New("!booyah")
		},
	}

	builder := New(client, "foo", 1)
	err := builder.Apply()

	require.NotNil(t, err)
	require.False(t, client.activateVersionCalled)
}

func Test_ActivateFailsWithError(t *testing.T) {

	client := &mockClient{
		cloneVersioner: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 1}, nil
		},
		activateVersioner: func(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
			return nil, errors.New("!booyah")
		},
	}

	builder := New(client, "foo", 1)
	err := builder.Apply()

	require.NotNil(t, err)

}

func Test_ActionCalledOnSuccess(t *testing.T) {

	client := &mockClient{
		cloneVersioner: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 1, ServiceID: "foo"}, nil
		},
		activateVersioner: func(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 2}, nil
		},
	}

	fn := func(current ServiceInfo) error {
		require.Equal(t, 1, current.Version)
		require.Equal(t, "foo", current.ID)
		return nil
	}

	builder := New(client, "foo", 1)
	err := builder.Apply(fn)

	require.Nil(t, err)
	require.True(t, client.cloneVersionCalled)
	require.True(t, client.activateVersionCalled)

}

func Test_ActionErrorPassedOn(t *testing.T) {

	client := &mockClient{
		cloneVersioner: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 1, ServiceID: "foo"}, nil
		},
		activateVersioner: func(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
			return &fastly.Version{Number: 2}, nil
		},
	}

	originalErr := errors.New("booyah")
	fn := func(current ServiceInfo) error {
		return originalErr
	}

	builder := New(client, "foo", 1)
	err := builder.Apply(fn)

	require.NotNil(t, err)
	require.Equal(t, originalErr, err)

}
