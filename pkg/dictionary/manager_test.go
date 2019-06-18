package dictionary

import (
	"testing"

	"github.com/fastly/go-fastly/fastly"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockRemoteSource struct {
	itemLister  func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	itemUpdater func(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	itemCreator func(*fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	itemDeleter func(i *fastly.DeleteDictionaryItemInput) error
}

func (m *mockRemoteSource) ListDictionaryItems(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return m.itemLister(i)
}
func (m *mockRemoteSource) UpdateDictionaryItem(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.itemUpdater(i)
}
func (m *mockRemoteSource) CreateDictionaryItem(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.itemCreator(i)
}
func (m *mockRemoteSource) DeleteDictionaryItem(i *fastly.DeleteDictionaryItemInput) error {
	return m.itemDeleter(i)
}

type mockLocalReader struct {
	reader func() (records [][]string, err error)
}

func (m *mockLocalReader) ReadAll() (records [][]string, err error) {
	return m.reader()
}

func Test_DiffAndMutate(t *testing.T) {
	testCases := []struct {
		name         string
		remoteLister func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
		local        func() (records [][]string, err error)
		created      int
		deleted      int
		updated      int
		err          error
	}{
		{
			name: "creations",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return [][]string{
					[]string{"one-key", "one-value"},
					[]string{"two-key", "two-value"},
				}, nil
			},
			created: 1,
		},
		{
			name: "deletions",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return nil, nil
			},
			deleted: 1,
		},
		{
			name: "updates",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return [][]string{
					[]string{"one-key", "foo"},
				}, nil
			},
			updated: 1,
		},
		{
			name: "no changes expected",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return [][]string{
					[]string{"one-key", "one-value"},
				}, nil
			},
		},
		{
			name: "all",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
					&fastly.DictionaryItem{ItemKey: "three-key", ItemValue: "three-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return [][]string{
					[]string{"one-key", "foo"},
					[]string{"two-key", "two-value"},
				}, nil
			},
			created: 1,
			updated: 1,
			deleted: 1,
		},
		{
			name: "duplicate locals fail",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			local: func() ([][]string, error) {
				return [][]string{
					[]string{"boo-key", "foo"},
					[]string{"boo-key", "bar"},
				}, nil
			},
			err: &DuplicateKeyErr{Key: "boo-key"},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			createCount := 0
			deleteCount := 0
			updateCount := 0

			client := &mockRemoteSource{
				itemCreator: func(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
					createCount++
					return &fastly.DictionaryItem{}, nil
				},
				itemDeleter: func(i *fastly.DeleteDictionaryItemInput) error {
					deleteCount++
					return nil
				},
				itemUpdater: func(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
					updateCount++
					return &fastly.DictionaryItem{}, nil
				},
				itemLister: tc.remoteLister,
			}

			local := &mockLocalReader{
				reader: tc.local,
			}

			m := Manager(client, WithLocalReader(local))

			err := m.Sync()

			if tc.err == nil {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Equal(t, tc.err, errors.Cause(err))
			}

			require.Equal(t, tc.created, createCount, "failed to create record(s)")
			require.Equal(t, tc.updated, updateCount, "failed to update record(s)")
			require.Equal(t, tc.deleted, deleteCount, "failed to delete record(s)")
		})
	}
}

func Test_RemoteDictionaryOption(t *testing.T) {

	service := "service-foo"
	dictionary := "dictionary-foo"
	count := 0

	client := &mockRemoteSource{
		itemCreator: func(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
			count++
			require.Equal(t, dictionary, i.Dictionary)
			require.Equal(t, service, i.Service)
			return &fastly.DictionaryItem{}, nil
		},
		itemDeleter: func(i *fastly.DeleteDictionaryItemInput) error {
			count++
			require.Equal(t, dictionary, i.Dictionary)
			require.Equal(t, service, i.Service)

			return nil
		},
		itemUpdater: func(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
			count++
			require.Equal(t, dictionary, i.Dictionary)
			require.Equal(t, service, i.Service)

			return &fastly.DictionaryItem{}, nil
		},
		itemLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
			count++
			require.Equal(t, dictionary, i.Dictionary)
			require.Equal(t, service, i.Service)

			return []*fastly.DictionaryItem{
				&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				&fastly.DictionaryItem{ItemKey: "three-key", ItemValue: "three-value"},
			}, nil
		},
	}

	local := &mockLocalReader{
		reader: func() ([][]string, error) {
			return [][]string{
				[]string{"one-key", "foo"},
				[]string{"two-key", "two-value"},
			}, nil
		},
	}

	m := Manager(client, WithLocalReader(local), WithRemoteDictionary(service, dictionary))

	err := m.Sync()

	require.Nil(t, err)
	require.Equal(t, 4, count)
}
