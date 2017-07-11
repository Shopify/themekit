package ystore

import (
	"testing"
)

func TestWrite(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()
	batch := store.Batch()

	if err := batch.Write("", "development", "one"); err == nil {
		t.Error("Allowed write of empty resource", err.Error())
	}
	if err := batch.Write(collection, "", "one"); err == nil {
		t.Error("Allowed write of empty key", err.Error())
	}
	if err := batch.Write(collection, "development", ""); err == nil {
		t.Error("Allowed write of empty value", err.Error())
	}
	if err = batch.Write(collection, "development", "development_time"); err != nil {
		t.Errorf("Failed to write: %v", err)
	}
	if batch.data[collection]["development"] != "development_time" {
		t.Errorf("Failed to write value")
	}

	batch = &Batch{}
	if err = batch.Write(collection, "development", "development_time"); err == nil {
		t.Errorf("batch didn't check for store")
	}
}

func TestCommit(t *testing.T) {
	store, err := New(storePath)
	if err != nil {
		t.Errorf("Unexpected error while creating store: %v", err)
	}
	defer store.Drop()

	batch := &Batch{}
	if err = batch.Commit(); err == nil {
		t.Errorf("batch didn't check for store")
	}

	if err := store.Write(collection, "development", "development_time"); err != nil {
		t.Errorf("failed to write %v", err)
	}

	batch = store.Batch()

	if err := batch.Write(collection, "development", "next_time"); err != nil {
		t.Errorf("failed to write %v", err)
	}

	if err := batch.Write("other", "development", "next_time"); err != nil {
		t.Errorf("failed to write %v", err)
	}

	if err = batch.Commit(); err != nil {
		t.Errorf("batch couldn't complete %v", err)
	}

	if store.data[collection]["development"] != "next_time" {
		t.Errorf("commit did not write value")
	}
}
