package dictionary

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/fastly"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	itemLister  func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	itemUpdater func(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	itemCreator func(*fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	itemDeleter func(i *fastly.DeleteDictionaryItemInput) error
}

func (m *mockClient) ListDictionaryItems(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return m.itemLister(i)
}
func (m *mockClient) UpdateDictionaryItem(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.itemUpdater(i)
}
func (m *mockClient) CreateDictionaryItem(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.itemCreator(i)
}
func (m *mockClient) DeleteDictionaryItem(i *fastly.DeleteDictionaryItemInput) error {
	return m.itemDeleter(i)
}

func Test_DiffAndMutate(t *testing.T) {
	testCases := []struct {
		name         string
		remoteLister func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
		localCSV     string
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
			localCSV: `one-key,one-value
two-key,two-value`,
			created: 1,
		},
		{
			name: "deletions",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			localCSV: ``,
			deleted:  1,
		},
		{
			name: "updates",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			localCSV: `one-key,foo`,
			updated:  1,
		},
		{
			name: "no changes",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			localCSV: `one-key,one-value`,
		},
		{
			name: "all",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
					&fastly.DictionaryItem{ItemKey: "three-key", ItemValue: "three-value"},
				}, nil
			},
			localCSV: `one-key,foo
two-key,two-value`,
			created: 1,
			updated: 1,
			deleted: 1,
		},
		{
			name: "duplicate-locals-fail",
			remoteLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
				return []*fastly.DictionaryItem{
					&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				}, nil
			},
			localCSV: `boo-key,one-value
boo-key,two-value`,
			err: &duplicateKeyErr{Key: "boo-key"},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			createCount := 0
			deleteCount := 0
			updateCount := 0

			client := &mockClient{
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

			reader := csv.NewReader(strings.NewReader(tc.localCSV))
			m := Manager(client, WithLocalCSVReader(reader))

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
