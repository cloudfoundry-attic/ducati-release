package store

import (
	"connet-api/models"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

const schema = `
CREATE TABLE IF NOT EXISTS routes (
  id bigserial primary key,
  app_guid text,
  fqdn text
);
`

//go:generate counterfeiter -o ../fakes/store.go --fake-name Store . Store
type Store interface {
	Create(route models.Route) error
	All() ([]models.Route, error)
}

//go:generate counterfeiter -o ../fakes/db.go --fake-name Db . db
type db interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type store struct {
	conn db
}

func New(connectionPool db) (Store, error) {
	_, err := connectionPool.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("setting up tables: %s", err)
	}

	return &store{
		conn: connectionPool,
	}, nil
}

func (s *store) Create(route models.Route) error {
	_, err := s.conn.NamedExec(
		`INSERT INTO routes (app_guid, fqdn) VALUES (:app_guid, :fqdn)`,
		&route,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			return fmt.Errorf("insert: %s", pqErr.Code.Name())
		}
		return fmt.Errorf("insert: %s", err)
	}

	return nil
}

func (s *store) All() ([]models.Route, error) {
	var routes []models.Route
	err := s.conn.Select(
		&routes,
		`SELECT app_guid, fqdn FROM routes`,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			return nil, fmt.Errorf("select: %s", pqErr.Code.Name())
		}
		return nil, fmt.Errorf("select: %s", err)
	}

	return routes, nil
}
