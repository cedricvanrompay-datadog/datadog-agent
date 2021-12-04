// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package uptane

import (
	"encoding/json"
	fmt "fmt"

	"github.com/DataDog/datadog-agent/pkg/config/remote/meta"
	"go.etcd.io/bbolt"
)

var (
	metaRootKey = []byte("root.json")
)

type localStore struct {
	metasBucket []byte
	rootsBucket []byte
	db          *bbolt.DB
}

func newLocalStore(db *bbolt.DB, repository string, cacheKey string, initialRoots meta.EmbeddedRoots) (*localStore, error) {
	s := &localStore{
		metasBucket: []byte(fmt.Sprintf("%s_%s_metas", repository, cacheKey)),
		rootsBucket: []byte(fmt.Sprintf("%s_%s_roots", repository, cacheKey)),
	}
	err := s.init(initialRoots)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *localStore) init(initialRoots meta.EmbeddedRoots) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(s.metasBucket)
		if err != nil {
			return fmt.Errorf("failed to create metas bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists(s.rootsBucket)
		if err != nil {
			return fmt.Errorf("failed to create roots bucket: %v", err)
		}
		rootsBucket := tx.Bucket(s.rootsBucket)
		for version, root := range initialRoots {
			rootKey := []byte(fmt.Sprintf("%d.root.json", version))
			err := rootsBucket.Put(rootKey, root)
			if err != nil {
				return fmt.Errorf("failed set embeded root in roots bucket: %v", err)
			}
		}
		metasBucket := tx.Bucket(s.rootsBucket)
		if metasBucket.Get(metaRootKey) == nil {
			err := rootsBucket.Put(metaRootKey, initialRoots.Last())
			if err != nil {
				return fmt.Errorf("failed set embeded root in roots bucket: %v", err)
			}
		}
		return nil
	})
}

// GetMeta returns a map of all the metadata files
func (s *localStore) GetMeta() (map[string]json.RawMessage, error) {
	meta := make(map[string]json.RawMessage)
	err := s.db.View(func(tx *bbolt.Tx) error {
		metaBucket := tx.Bucket(s.metasBucket)
		cursor := metaBucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			tmp := make([]byte, len(v))
			copy(tmp, v)
			meta[string(k)] = json.RawMessage(tmp)
		}
		return nil
	})
	return meta, err
}

// SetMeta stores a tuf metadata file
func (s *localStore) SetMeta(name string, meta json.RawMessage) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		metaBucket := tx.Bucket(s.metasBucket)
		return metaBucket.Put([]byte(name), meta)
	})
}

// DeleteMeta deletes a tuf metadata file
func (s *localStore) DeleteMeta(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		metaBucket := tx.Bucket(s.metasBucket)
		return metaBucket.Delete([]byte(name))
	})
}

type localStoreDirector struct {
	*localStore
}

func newLocalStoreDirector(db *bbolt.DB, cacheKey string) (*localStoreDirector, error) {
	localStore, err := newLocalStore(db, "director", cacheKey, meta.RootsDirector())
	if err != nil {
		return nil, err
	}
	return &localStoreDirector{
		localStore: localStore,
	}, nil
}

type localStoreConfig struct {
	*localStore
}

func newLocalStoreConfig(db *bbolt.DB, cacheKey string) (*localStoreConfig, error) {
	localStore, err := newLocalStore(db, "config", cacheKey, meta.RootsDirector())
	if err != nil {
		return nil, err
	}
	return &localStoreConfig{
		localStore: localStore,
	}, nil
}