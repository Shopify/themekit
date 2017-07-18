package ystore

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v1"
)

// YStore is the struct that handles all interaction with the datastore.
type YStore struct {
	mutex   sync.Mutex
	path    string
	data    map[string]map[string]string
	comment string
}

var (
	defaultPerms  os.FileMode = 0644
	currentStores             = make(map[string]*YStore)

	// ErrorInvalidCollectionName is returned when a collection param is an empty string
	ErrorInvalidCollectionName = errors.New("No collection name provided")
	// ErrorInvalidKeyName is returned when a key param is an empty string
	ErrorInvalidKeyName = errors.New("No key name provided")
	// ErrorInvalidValue is returned when a key param is an empty string
	ErrorInvalidValue = errors.New("No value provided, if you are trying to delete the key please use Delete()")
	// ErrorCollectionNotFound is returned when trying to read a non exitant collection
	ErrorCollectionNotFound = errors.New("collection not found")
	// ErrorKeyNotFound is returned when trying to read a non exitant key
	ErrorKeyNotFound = errors.New("key not found")
)

// New will create and return a new YStore. A error will be returned if the
// store already exists and contains invalid data.
func New(path string) (*YStore, error) {
	if _, ok := currentStores[path]; !ok {
		currentStores[path] = &YStore{
			path: filepath.Clean(path),
			data: make(map[string]map[string]string),
		}
	}
	return currentStores[path], currentStores[path].read()
}

// SetComment will set a comment that will be prepended to the file.
func (store *YStore) SetComment(comment string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.comment = comment
	return store.flush()
}

// Batch will return a batch object for this store that can be used to write multiple
// values ona single file write.
func (store *YStore) Batch() *Batch {
	return &Batch{
		store: store,
		data:  make(map[string]map[string]string),
	}
}

// Write will set the value for a key in a collection. Write also lazily instantiates
// the store so that if the file does not already exist, Write will create it. Write
// requires that all strings provided to it are not blank.
func (store *YStore) Write(collection, key, value string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if collection == "" {
		return ErrorInvalidCollectionName
	} else if key == "" {
		return ErrorInvalidKeyName
	} else if value == "" {
		return ErrorInvalidValue
	}
	if _, ok := store.data[collection]; !ok {
		store.data[collection] = make(map[string]string)
	}
	store.data[collection][key] = value
	return store.flush()
}

// Read will read and return the value from the datastore for the provided
// collection key. Read requires that the collection and key are not empty
// strings otherwise an error will be returned. If the collection or key does
// not exist, an error will be returned.
func (store *YStore) Read(collection, key string) (string, error) {
	if collection == "" {
		return "", ErrorInvalidCollectionName
	} else if key == "" {
		return "", ErrorInvalidKeyName
	}

	if err := store.read(); err != nil {
		return "", err
	}

	if _, ok := store.data[collection]; !ok {
		return "", ErrorCollectionNotFound
	}

	val, ok := store.data[collection][key]
	if !ok {
		return "", ErrorKeyNotFound
	}

	return val, nil
}

// Collections will return a string array of all the collections in the datastore.
// Collections will only ever return an error if the datastore is corrupt.
func (store *YStore) Collections() ([]string, error) {
	if err := store.read(); err != nil {
		return []string{}, err
	}

	keys := make([]string, len(store.data))
	i := 0
	for key := range store.data {
		keys[i] = key
		i++
	}

	return keys, nil
}

// Dump will return the total contents of the store
func (store *YStore) Dump() (map[string]map[string]string, error) {
	if err := store.read(); err != nil {
		return nil, err
	}
	return store.data, nil
}

// ReadAll will return a string array of all the keys in a collections. ReadAll
// will return an error if the collection does not exist or the collection parameter
// is an empty string.
func (store *YStore) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return []string{}, ErrorInvalidCollectionName
	}

	if err := store.read(); err != nil {
		return []string{}, err
	}

	if _, ok := store.data[collection]; !ok {
		return []string{}, ErrorCollectionNotFound
	}

	keys := make([]string, len(store.data[collection]))
	i := 0
	for key := range store.data[collection] {
		keys[i] = key
		i++
	}

	return keys, nil
}

// Delete will remove a key/value from the collection in the data store. An error
// will be returned if the collection cannot be found or the change cannot be persisted.
func (store *YStore) Delete(collection, key string) error {
	if collection == "" {
		return ErrorInvalidCollectionName
	} else if key == "" {
		return ErrorInvalidKeyName
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.data[collection]; !ok {
		return ErrorCollectionNotFound
	}

	delete(store.data[collection], key)
	if len(store.data[collection]) == 0 {
		delete(store.data, collection)
	}
	return store.flush()
}

// DeleteCollection will remove the collection from the database. DeleteCollection
// will return an error if the change cannot be persisted.
func (store *YStore) DeleteCollection(collection string) error {
	if collection == "" {
		return ErrorInvalidCollectionName
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.data, collection)
	return store.flush()
}

// Drop will clear the data cache as well as remove the file on the disk that acts
// as the datastore. This will completely clear the datastore.
func (store *YStore) Drop() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.data = make(map[string]map[string]string)
	return os.Remove(store.path)
}

func (store *YStore) read() error {
	if _, err := os.Stat(store.path); err != nil {
		return nil
	}

	contents, err := ioutil.ReadFile(store.path)
	if err != nil {
		return err
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	if err = yaml.Unmarshal(contents, &store.data); err != nil {
		return err
	}

	return nil
}

func (store *YStore) flush() error {
	file, err := os.OpenFile(store.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultPerms)
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := yaml.Marshal(store.data)
	if err != nil {
		return err
	}
	if _, err = file.Write(append([]byte("# "+store.comment+"\n"), bytes...)); err != nil {
		return err
	}
	return nil
}
