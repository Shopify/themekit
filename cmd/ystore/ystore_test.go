package ystore

import (
	"os"
	"testing"
)

var (
	storePath  = "./mystore"
	collection = "assets/application.js"
)

func TestNew(t *testing.T) {
	if _, err := os.Stat(storePath); err == nil {
		t.Error("Expected nothing, got file path")
	}
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	if _, err = os.Stat(storePath); err == nil {
		t.Error("Expected nothing, got store")
	}
	if newStore, err := New(storePath); err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	} else if newStore != store {
		t.Errorf("new store is not same instance as old")
	}
	if err = store.Write(collection, "development", "development_time"); err != nil {
		t.Errorf("Failed to write: %v", err)
	}
	if _, err = os.Stat(storePath); err != nil {
		t.Error("Expected store, got nothing")
	}
}

func TestSetComment(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	store.SetComment("This is comment")
	if store.comment != "This is comment" {
		t.Errorf("wasn't able to set comment")
	}
}

func TestWriteAndRead(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	if err = store.Write(collection, "development", "development_time"); err != nil {
		t.Errorf("Failed to write: %v", err)
	}
	var val string
	if val, err = store.Read(collection, "development"); err != nil {
		t.Error("Failed to read: ", err.Error())
	}
	if val != "development_time" {
		t.Errorf("bad value while reading: %v", val)
	}
}

func TestDump(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	if err = store.Write(collection, "development", "development_time"); err != nil {
		t.Errorf("Failed to write: %v", err)
	}

	dump, err := store.Dump()
	if err != nil {
		t.Errorf("Failed to dump: %v", err)
	}
	if dump[collection]["development"] != "development_time" {
		t.Errorf("dump didn't have appropriate data")
	}
}

func TestRead(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()

	store.Write(collection, "development", "one")
	store.Write(collection, "production", "two")

	if keys, err := store.ReadAll(collection); err != nil {
		t.Errorf("Failed to read: %v", err)
	} else if len(keys) != 2 {
		t.Error("Expected some keys, have none")
	}

	if collections, err := store.Collections(); err != nil {
		t.Errorf("Failed to read: %v", err)
	} else if len(collections) != 1 || collections[0] != collection {
		t.Errorf("Failed to read collections")
	}

	val, err := store.Read(collection, "no")
	if val != "" || err != ErrorKeyNotFound {
		t.Errorf("didn't return errorkeynotfound")
	}
}

func TestWriteAndReadEmpty(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	if err := store.Write("", "development", "one"); err == nil {
		t.Error("Allowed write of empty resource", err.Error())
	}
	if err := store.Write(collection, "", "one"); err == nil {
		t.Error("Allowed write of empty resource", err.Error())
	}
	if err := store.Write(collection, "development", ""); err == nil {
		t.Error("Allowed write of empty value", err.Error())
	}
	if _, err := store.Read("", "development"); err == nil {
		t.Error("Read: Allowed read of empty resource", err.Error())
	}
	if _, err := store.Read(collection, ""); err == nil {
		t.Error("Read: Allowed read of empty resource", err.Error())
	}
	if _, err := store.Read(collection, "nope"); err == nil {
		t.Error("Read: Allowed read of non existent collection", err.Error())
	}
	if _, err := store.ReadAll(""); err == nil {
		t.Error("ReadAll: Allowed read of empty resource", err.Error())
	}
	if _, err := store.ReadAll("nope"); err == nil {
		t.Error("ReadAll: Allowed read of non existent collection", err.Error())
	}
}

func TestDelete(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	if err := store.Write(collection, "development", "one"); err != nil {
		t.Errorf("Create collection failed: %v", err)
	}
	if err := store.Delete(collection, "development"); err != nil {
		t.Errorf("Failed to delete: %v", err)
	}
	if _, ok := store.data[collection]; ok {
		t.Errorf("delete did not also delete collection when empty")
	}
	if err := store.Delete("", "development"); err == nil {
		t.Errorf("Allowed empty collection to delete: %v", err)
	}
	if err := store.Delete(collection, ""); err == nil {
		t.Errorf("Allowed empty key to delete: %v", err)
	}
	if err := store.Delete("nope", "development"); err == nil {
		t.Errorf("Allowed delete on non existent collection: %v", err)
	}
	if _, err := store.Read(collection, "development"); err == nil {
		t.Error("Expected nothing, got value")
	}
}

func TestDeleteCollection(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	store.Write(collection, "development", "one")
	store.Write(collection, "production", "two")
	if err := store.DeleteCollection(collection); err != nil {
		t.Error("Failed to delete: ", err.Error())
	}
	if err := store.DeleteCollection(""); err == nil {
		t.Errorf("Allowed empty collection to delete: %v", err)
	}
}

func TestDrop(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	store.Write(collection, "development", "one")
	if _, err = os.Stat(storePath); err != nil {
		t.Error("Expected store, got nothing")
	}
	store.Drop()
	if _, err = os.Stat(storePath); err == nil {
		t.Error("Expected nothing, got store")
	}
}
