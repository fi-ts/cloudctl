package mssql

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Store struct {
	databases []Database
}

func seedDatabases() []Database {
	const seedTimestamp = "2024-01-01T00:00:00Z"
	return []Database{
		{
			ID:               "mssql-db-001",
			Name:             "orders-prod",
			Project:          "project-alpha",
			Version:          "2022",
			StorageGB:        100,
			Host:             "mssql-db-001.example.internal",
			Port:             DefaultPort,
			Status:           StatusRunning,
			CreatedTimestamp: seedTimestamp,
		},
		{
			ID:               "mssql-db-002",
			Name:             "analytics",
			Project:          "project-beta",
			Version:          "2019",
			StorageGB:        250,
			Host:             "mssql-db-002.example.internal",
			Port:             DefaultPort,
			Status:           StatusRunning,
			CreatedTimestamp: seedTimestamp,
		},
		{
			ID:               "mssql-db-003",
			Name:             "staging-users",
			Project:          "project-alpha",
			Version:          "2022",
			StorageGB:        50,
			Host:             "mssql-db-003.example.internal",
			Port:             DefaultPort,
			Status:           StatusStopped,
			CreatedTimestamp: seedTimestamp,
		},
	}
}

func NewStore() *Store {
	return &Store{databases: seedDatabases()}
}

func (s *Store) List() ([]Database, error) {
	out := make([]Database, len(s.databases))
	copy(out, s.databases)
	return out, nil
}

func (s *Store) Create(name, version, project string, storageGB int) (*Database, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateVersion(version); err != nil {
		return nil, err
	}
	if err := validateStorageGB(storageGB); err != nil {
		return nil, err
	}

	db := Database{
		ID:               s.newID(),
		Name:             name,
		Project:          project,
		Version:          version,
		StorageGB:        storageGB,
		Host:             "",
		Port:             DefaultPort,
		Status:           StatusPending,
		CreatedTimestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.databases = append(s.databases, db)

	created := db
	return &created, nil
}

func (s *Store) Delete(id, project string) error {
	for i, db := range s.databases {
		if db.ID == id && db.Project == project {
			s.databases = append(s.databases[:i], s.databases[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("database %q not found in project %q", id, project)
}

func (s *Store) newID() string {
	for {
		id := "mssql-db-" + randomHex(6)
		if !s.hasID(id) {
			return id
		}
	}
}

func (s *Store) hasID(id string) bool {
	for _, db := range s.databases {
		if db.ID == id {
			return true
		}
	}
	return false
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
