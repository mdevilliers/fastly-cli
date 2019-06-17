package dictionary

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/fastly"
	"github.com/stretchr/testify/assert"
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

func Test_AdditionalRecordsCreated(t *testing.T) {

	createCount := 0

	client := &mockClient{
		itemLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
			return []*fastly.DictionaryItem{
				&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
			}, nil
		},
		itemCreator: func(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
			createCount++
			//fmt.Println(i.ItemKey, i.ItemValue)
			return &fastly.DictionaryItem{}, nil
		},
	}
	local := `one-key,one-value
two-key,two-value`
	reader := csv.NewReader(strings.NewReader(local))

	m := Manager(client, WithLocalCSVReader(reader))

	err := m.Sync()
	require.Nil(t, err)
	assert.Equal(t, 1, createCount)
}

func Test_RemovedRecordsDeleted(t *testing.T) {

	deleteCount := 0

	client := &mockClient{
		itemLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
			return []*fastly.DictionaryItem{
				&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				&fastly.DictionaryItem{ItemKey: "two-key", ItemValue: "two-value"},
			}, nil
		},
		itemDeleter: func(i *fastly.DeleteDictionaryItemInput) error {
			deleteCount++
			//fmt.Println(i.ItemKey, i.ItemValue)
			return nil
		},
	}
	local := `one-key,one-value`
	reader := csv.NewReader(strings.NewReader(local))

	m := Manager(client, WithLocalCSVReader(reader))

	err := m.Sync()
	require.Nil(t, err)
	assert.Equal(t, 1, deleteCount)
}

func Test_ChangedRecordsUpdated(t *testing.T) {

	updateCount := 0

	client := &mockClient{
		itemLister: func(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
			return []*fastly.DictionaryItem{
				&fastly.DictionaryItem{ItemKey: "one-key", ItemValue: "one-value"},
				&fastly.DictionaryItem{ItemKey: "two-key", ItemValue: "two-value"},
			}, nil
		},
		itemUpdater: func(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
			updateCount++
			//fmt.Println(i.ItemKey, i.ItemValue)
			return &fastly.DictionaryItem{}, nil
		},
	}
	local := `one-key,foo-value
two-key,bar-value`

	reader := csv.NewReader(strings.NewReader(local))

	m := Manager(client, WithLocalCSVReader(reader))

	err := m.Sync()
	require.Nil(t, err)
	assert.Equal(t, 2, updateCount)
}
