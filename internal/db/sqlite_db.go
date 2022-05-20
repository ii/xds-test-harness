package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/internal/parser"
	"github.com/mattn/go-sqlite3"
)

var (
	ErrDuplicate    = errors.New("Record already exists")
	ErrNotExists    = errors.New("Record does not exist")
	ErrUpdateFailed = errors.New("Update failed")
	ErrDeleteFailed = errors.New("Delete failed")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSqliteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (s *SQLiteRepository) Migrate() error {
	_, err := s.db.Exec(MigrateSQL)
	return err
}

func (s *SQLiteRepository) InsertRequest(req *discovery.DiscoveryRequest) error {
	b, err := parser.ProtoJSONMarshal(req)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(InsertRequestSQL, string(b))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return err
			}
		}
		return err
	}
	return err
}

func (s *SQLiteRepository) InsertResponse(res *discovery.DiscoveryResponse) error {
	b, err := parser.ProtoJSONMarshal(res)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(InsertResponseSQL, string(b))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return err
			}
		}
		return err
	}
	return err
}

func (s *SQLiteRepository) DeleteAll() error {
	_, err := s.db.Exec(DeleteAllSQL)
	if err != nil {
		return fmt.Errorf("Couldn't delete the records: %v", err)
	}
	return nil
}

func (s *SQLiteRepository) CheckExpectedResources(resources []string, version, typeUrl string) (match bool, single bool, err error) {
	var resources_match int64
	var single_response int64
	r, err := json.Marshal(resources)
	if err != nil {
		return false, false, fmt.Errorf("Had issue turning resource into valid json. Cannot run a db query: %v", err)
	}
	row := s.db.QueryRow(CheckExpectedResourcesSQL, r, version, typeUrl)
	if err := row.Scan(&resources_match, &single_response); err != nil {
		return false, false, err
	}
	return resources_match != 0, single_response != 0, err
}

func (s *SQLiteRepository) CheckOnlyExpectedResources(resources []string, version, typeUrl string) (passed bool, err error) {
	var valid int64
	r, err := json.Marshal(resources)
	if err != nil {
		return false, fmt.Errorf("Had issue turning resource into valid json. Cannot run a db query: %v", err)
	}
	row := s.db.QueryRow(CheckOnlyExpectedResourcesSQL, r, version, typeUrl)
	if err := row.Scan(&valid); err != nil {
		return false, err
	}
	return valid != 0, err
}

func (s *SQLiteRepository) DeltaCheckOnlyExpectedResources(resources []string, version, typeUrl string) (passed bool, err error) {
	var valid int64
	r, err := json.Marshal(resources)
	if err != nil {
		return false, fmt.Errorf("Had issue turning resource into valid json. Cannot run a db query: %v", err)
	}
	row := s.db.QueryRow(DeltaCheckOnlyExpectedResourcesSQL, r, version, typeUrl)
	if err := row.Scan(&valid); err != nil {
		return false, err
	}
	return valid != 0, err
}

func (s *SQLiteRepository) CheckMoreRequestsThanResponses() (bool, error) {
	var check int64
	row := s.db.QueryRow(CheckMoreRequestsThanResponseSQL)
	if err := row.Scan(&check); err != nil {
		return false, err
	}
	return check != 0, nil
}

func (s *SQLiteRepository) CheckNoResponsesForVersion(version string) (bool, error) {
	var check int64
	row := s.db.QueryRow(CheckNoResponsesForVersionSQL, version)
	if err := row.Scan(&check); err != nil {
		return false, err
	}
	return check != 0, nil
}
