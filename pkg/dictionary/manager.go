package dictionary

import (
	"fmt"

	"github.com/fastly/go-fastly/fastly"
	"github.com/pkg/errors"
	"github.com/r3labs/diff"
)

type manager struct {
	serviceID  string
	dictionary string
	local      localReader
	remote     remoteDictionaryMutator
}

type option func(*manager)

// WithRemoteDictionary allows specifying the Fastly service and dictionary to use
func WithRemoteDictionary(serviceID, dictionary string) option {
	return func(m *manager) {
		m.serviceID = serviceID
		m.dictionary = dictionary
	}
}

type localReader interface {
	ReadAll() (records [][]string, err error)
}

// WithLocalReader allows specifying the local dictionary provider
func WithLocalReader(reader localReader) option {
	return func(m *manager) {
		m.local = reader
	}
}

type remoteDictionaryMutator interface {
	ListDictionaryItems(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	UpdateDictionaryItem(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	CreateDictionaryItem(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	DeleteDictionaryItem(i *fastly.DeleteDictionaryItemInput) error
}

// Manager returns a way of syncing a local dictionary with a remote one
func Manager(client remoteDictionaryMutator, options ...option) *manager { // nolint
	m := &manager{
		remote: client,
	}

	for _, o := range options {
		o(m)
	}

	return m
}

// Sync syncs a local dictionary with a remote one or returns an error
// Local items not remotely available are added
// Remote items not locally available are deleted
// Changed local items are updated
func (m *manager) Sync() error {

	// get all or the remote items
	remoteItems, err := m.remote.ListDictionaryItems(&fastly.ListDictionaryItemsInput{
		Service:    m.serviceID,
		Dictionary: m.dictionary,
	})

	if err != nil {
		return errors.Wrap(err, "error retrieving dictionary items")
	}

	localItems, err := m.local.ReadAll()

	if err != nil {
		return errors.Wrap(err, "error reading local dictionary items")
	}

	changelog, err := m.diff(remoteItems, localItems)

	if err != nil {
		return errors.Wrap(err, "error diffing remote and local dictionary items")
	}

	for change := range changelog {

		key := changelog[change].Path[0]

		if changelog[change].Type == diff.CREATE {

			value := changelog[change].To.(string)
			_, err := m.remote.CreateDictionaryItem(&fastly.CreateDictionaryItemInput{
				Service:    m.serviceID,
				Dictionary: m.dictionary,
				ItemKey:    key,
				ItemValue:  value,
			})

			if err != nil {
				return errors.Wrapf(err, "error adding item : %s", key)
			}

		}
		if changelog[change].Type == diff.DELETE {
			err := m.remote.DeleteDictionaryItem(&fastly.DeleteDictionaryItemInput{
				Service:    m.serviceID,
				Dictionary: m.dictionary,
				ItemKey:    key,
			})

			if err != nil {
				return errors.Wrapf(err, "error deleting item : %s", key)
			}

		}
		if changelog[change].Type == diff.UPDATE {

			value := changelog[change].To.(string)
			_, err := m.remote.UpdateDictionaryItem(&fastly.UpdateDictionaryItemInput{
				Service:    m.serviceID,
				Dictionary: m.dictionary,
				ItemKey:    key,
				ItemValue:  value,
			})

			if err != nil {
				return errors.Wrapf(err, "error updating item : %s", key)
			}
		}
	}
	return nil
}

func (m *manager) diff(remote []*fastly.DictionaryItem, local [][]string) (diff.Changelog, error) {

	localMap, err := stringSliceSliceToMap(local)

	if err != nil {
		return nil, err
	}

	remoteMap := fastlyDictionaryItemsToMap(remote)

	return diff.Diff(remoteMap, localMap)
}

func fastlyDictionaryItemsToMap(a []*fastly.DictionaryItem) map[string]string {

	m := map[string]string{}

	for i := range a {
		k := a[i].ItemKey
		v := a[i].ItemValue
		m[k] = v
	}

	return m
}

func stringSliceSliceToMap(a [][]string) (map[string]string, error) {

	m := map[string]string{}

	for i := range a {
		k := a[i][0]
		v := a[i][1]

		_, contains := m[k]
		if contains {
			return nil, &DuplicateKeyErr{Key: k}
		}

		m[k] = v
	}
	return m, nil

}

// DuplicateKeyErr captures the offending key that exists more then once locally
type DuplicateKeyErr struct {
	Key string
}

func (d *DuplicateKeyErr) Error() string {
	return fmt.Sprintf("duplicate key : %s", d.Key)
}
