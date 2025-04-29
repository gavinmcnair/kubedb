package store

import (
	"github.com/dgraph-io/badger/v3"
	"sync"
	"errors"
)

// BadgerDBStore represents a key-value store backed by BadgerDB.
type BadgerDBStore struct {
	db       *badger.DB
	mu       sync.RWMutex // Mutex for synchronizing access
}

// NewBadgerDBStore initializes a new BadgerDB store at the given path.
func NewBadgerDBStore(path string) (*BadgerDBStore, error) {
	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDBStore{
		db: db,
	}, nil
}

// Put stores a key-value pair in the database.
func (store *BadgerDBStore) Put(key, value []byte) error {
	return store.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves the value associated with the given key.
func (store *BadgerDBStore) Get(key []byte) ([]byte, error) {
	var itemValue []byte
	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		itemValue, err = item.ValueCopy(nil)
		return err
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.New("key not found")
	}
	return itemValue, err
}

// Delete removes the key-value pair associated with the given key.
func (store *BadgerDBStore) Delete(key []byte) error {
	return store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close shuts down the database connection.
func (store *BadgerDBStore) Close() {
	store.db.Close()
}

