package dictionary

import (
	"fmt"
	"net/http"

	"github.com/fastly/go-fastly/fastly"
	fastlyext "github.com/mdevilliers/fastly-cli/pkg/fastly-ext"
	"github.com/pkg/errors"
	"github.com/r3labs/diff"
)

const (
	// https://docs.fastly.com/guides/edge-dictionaries/about-edge-dictionaries
	// Dictionaries are limited to 1000 items.
	maxItems = 1000
	// Dictionary item keys are limited to 256 characters and their values are limited to 8000 characters
	maxKeyLength   = 256
	maxValueLength = 8000
)

var (
	// ErrTooManyItems signals the Fastly maximum items has been reached
	ErrTooManyItems = errors.New("too many items")
)

type manager struct {
	serviceID    string
	dictionaryID string
	local        localReader
	remote       remoteDictionaryMutator
}

type option func(*manager)

// WithRemoteDictionary allows specifying the Fastly service and dictionary to use
// NOTE : this that function requires ID's and NOT the name's of the entities
func WithRemoteDictionary(serviceID, dictionaryID string) option {
	return func(m *manager) {
		m.serviceID = serviceID
		m.dictionaryID = dictionaryID
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
	BatchUpdateDictionaryItems(i *fastlyext.BatchUpdateDictionaryItemsInput) error
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
		Dictionary: m.dictionaryID,
	})

	if err != nil {

		httpError, ok := err.(*fastly.HTTPError)

		if ok {
			if httpError.StatusCode == http.StatusNotFound {
				return errors.New("dictionary not found")
			}
		}

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

	batchUpdates := []fastlyext.BatchUpdateDictionaryItem{}

	for change := range changelog {

		key := changelog[change].Path[0]

		if changelog[change].Type == diff.CREATE {

			value := changelog[change].To.(string)

			batchUpdates = append(batchUpdates, fastlyext.BatchUpdateDictionaryItem{
				Operation: fastlyext.CreateBatchOperation,
				Key:       key,
				Value:     value,
			})

		}

		if changelog[change].Type == diff.DELETE {

			batchUpdates = append(batchUpdates, fastlyext.BatchUpdateDictionaryItem{
				Operation: fastlyext.DeleteBatchOperation,
				Key:       key,
			})

		}

		if changelog[change].Type == diff.UPDATE {

			value := changelog[change].To.(string)

			batchUpdates = append(batchUpdates, fastlyext.BatchUpdateDictionaryItem{
				Operation: fastlyext.UpdateBatchOperation,
				Key:       key,
				Value:     value,
			})

		}

		// 1000 is the maximum batch size
		// If we have reached this amount flush the batch now
		if len(batchUpdates) == fastlyext.BatchUpdateMaximumItems {

			err := m.remote.BatchUpdateDictionaryItems(&fastlyext.BatchUpdateDictionaryItemsInput{
				Service:    m.serviceID,
				Dictionary: m.dictionaryID,
				Items:      batchUpdates,
			})

			if err != nil {
				return errors.Wrap(err, "error updating batch")
			}
			batchUpdates = []fastlyext.BatchUpdateDictionaryItem{}
		}
	}

	if len(batchUpdates) == 0 {
		return nil
	}

	// flush the last batch if any
	return m.remote.BatchUpdateDictionaryItems(&fastlyext.BatchUpdateDictionaryItemsInput{
		Service:    m.serviceID,
		Dictionary: m.dictionaryID,
		Items:      batchUpdates,
	})

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

	if len(a) > maxItems {
		return nil, ErrTooManyItems
	}

	m := map[string]string{}

	for i := range a {
		k := a[i][0]
		v := a[i][1]

		_, contains := m[k]
		if contains {
			return nil, &ErrDuplicateKey{Key: k}
		}

		err := validateItem(k, v)

		if err != nil {
			return nil, err
		}

		m[k] = v
	}
	return m, nil

}

func validateItem(key, value string) error {

	if len(key) > maxKeyLength {
		return &ErrKeyTooLong{Key: key}
	}
	if len(value) > maxValueLength {
		return &ErrValueTooLong{Key: key, Value: value}
	}

	return nil
}

// ErrDuplicateKey captures the offending key that exists more then once locally
type ErrDuplicateKey struct {
	Key string
}

func (d *ErrDuplicateKey) Error() string {
	return fmt.Sprintf("duplicate key : %s", d.Key)
}

// ErrKeyTooLong signals the Key is too long to be stored
type ErrKeyTooLong struct {
	Key string
}

func (k *ErrKeyTooLong) Error() string {
	return fmt.Sprintf("key too long (max : %v) : %s", maxKeyLength, k.Key)
}

// ErrValueTooLong signals the Value is too long to be stored
type ErrValueTooLong struct {
	Key   string
	Value string
}

func (v *ErrValueTooLong) Error() string {
	return fmt.Sprintf("value too long (max : %v) : %s : %s", maxValueLength, v.Key, v.Value)
}
