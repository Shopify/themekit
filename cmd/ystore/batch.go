package ystore

import (
	"fmt"
	"sync"
)

// Batch handles multiple writes without a write to the disk so that many values
// can be written at a single time
type Batch struct {
	mutex sync.Mutex
	store *YStore
	data  map[string]map[string]string
}

// Write will set a value on a collection key in the batch. This will have no
// effect on the store until Commit() is called on the batch
func (batch *Batch) Write(collection, key, value string) error {
	if batch.store == nil {
		return fmt.Errorf("No store set on batch")
	}

	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	if collection == "" {
		return ErrorInvalidCollectionName
	} else if key == "" {
		return ErrorInvalidKeyName
	} else if value == "" {
		return ErrorInvalidValue
	}
	if _, ok := batch.data[collection]; !ok {
		batch.data[collection] = make(map[string]string)
	}
	batch.data[collection][key] = value
	return nil
}

// Commit will lock the batch and store while it writes all of the values at one
// time to the data store
func (batch *Batch) Commit() error {
	if batch.store == nil {
		return fmt.Errorf("No store set on batch")
	}

	batch.mutex.Lock()
	batch.store.mutex.Lock()
	defer batch.mutex.Unlock()
	defer batch.store.mutex.Unlock()
	for colName, collection := range batch.data {
		for key, value := range collection {
			if _, ok := batch.store.data[colName]; !ok {
				batch.store.data[colName] = make(map[string]string)
			}
			batch.store.data[colName][key] = value
		}
	}
	if err := batch.store.flush(); err != nil {
		return err
	}
	batch.data = make(map[string]map[string]string)
	return nil
}
