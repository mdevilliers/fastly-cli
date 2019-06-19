package fastlyext

import (
	"errors"
	"fmt"

	"github.com/fastly/go-fastly/fastly"
)

// BatchUpdateDictionaryItemsInput is used as an input to the  BatchUpdateDictionaryItems function
type BatchUpdateDictionaryItemsInput struct {
	// Service is the ID of the service. Dictionary is the ID of the dictionary.
	// Both fields are required.
	Service    string
	Dictionary string

	Items []BatchUpdateDictionaryItem `json:"items"`
}

// BatchOperation is the operation for the update
type BatchOperation string

var (
	// CreateBatchOperation creates the dictionary item
	CreateBatchOperation = BatchOperation("create")

	// UpdateBatchOperation updates the existing dictionary item
	UpdateBatchOperation = BatchOperation("update")

	// UpsertBatchOperation upserts the existing dictionary item
	UpsertBatchOperation = BatchOperation("upsert")

	// DeleteBatchOperation deletes the existing dictionary item
	DeleteBatchOperation = BatchOperation("delete")

	// ErrBatchUpdateMaximumItemsExceeded signals the user has supplied more than the maximum
	// number of items as specified by Fastly
	ErrBatchUpdateMaximumItemsExceeded = errors.New("batch update maximum items exceeded")
)

const (
	BatchUpdateMaximumItems = 1000
)

// BatchUpdateDictionaryItem holds the information for the update
type BatchUpdateDictionaryItem struct {
	Operation BatchOperation `json:"op"`
	Key       string         `json:"item_key"`
	Value     string         `json:"item_value"`
}

// BatchUpdateDictionaryItems performs the update or returns an error
func (c *Client) BatchUpdateDictionaryItems(i *BatchUpdateDictionaryItemsInput) error {
	if i.Service == "" {
		return fastly.ErrMissingService
	}

	if i.Dictionary == "" {
		return fastly.ErrMissingDictionary
	}

	if len(i.Items) > BatchUpdateMaximumItems {
		return ErrBatchUpdateMaximumItemsExceeded
	}

	path := fmt.Sprintf("/service/%s/dictionary/%s/items", i.Service, i.Dictionary)
	resp, err := c.PatchJSON(path, i, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint

	// This endpoint returns an object with a status of 'ok' according to the documentation
	// Other responses are undocumented.
	// It is reasonable to rely on the HTTP status.
	return nil
}
