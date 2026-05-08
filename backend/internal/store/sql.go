package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type sqlRunner interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

type txRunner interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

var namedParamPattern = regexp.MustCompile(`:[A-Za-z_][A-Za-z0-9_]*`)

func (s *Store) exec(query string, args ...any) (sql.Result, error) {
	return s.execWith(s.db, query, args...)
}

func (s *Store) query(query string, args ...any) (*sql.Rows, error) {
	return s.queryWith(s.db, query, args...)
}

func (s *Store) queryRow(query string, args ...any) *sql.Row {
	query, args = s.prepareQuery(query, args)
	return s.db.QueryRow(query, args...)
}

func (s *Store) prepare(query string) (*sql.Stmt, error) {
	return s.prepareWith(s.db, query)
}

func (s *Store) execWith(r sqlRunner, query string, args ...any) (sql.Result, error) {
	query, args = s.prepareQuery(query, args)
	return r.Exec(query, args...)
}

func (s *Store) queryWith(r sqlRunner, query string, args ...any) (*sql.Rows, error) {
	query, args = s.prepareQuery(query, args)
	return r.Query(query, args...)
}

func (s *Store) queryRowWith(r sqlRunner, query string, args ...any) *sql.Row {
	query, args = s.prepareQuery(query, args)
	return r.QueryRow(query, args...)
}

func (s *Store) prepareWith(r interface {
	Prepare(string) (*sql.Stmt, error)
}, query string) (*sql.Stmt, error) {
	query, _ = s.prepareQuery(query, nil)
	return r.Prepare(query)
}

func (s *Store) insertAndReturnID(r sqlRunner, query string, args ...any) (int64, error) {
	var id int64
	err := s.queryRowWith(r, query+" RETURNING id", args...).Scan(&id)
	return id, err
}

func (s *Store) prepareQuery(query string, args []any) (string, []any) {
	query = strings.ReplaceAll(query, "unixepoch()", "CAST(EXTRACT(EPOCH FROM NOW()) AS BIGINT)")
	query = strings.ReplaceAll(query, " LIKE :query", " ILIKE :query")

	if len(args) == 0 {
		return query, args
	}

	namedArgs := map[string]any{}
	for _, arg := range args {
		if named, ok := arg.(sql.NamedArg); ok {
			namedArgs[named.Name] = named.Value
		}
	}
	if len(namedArgs) == 0 {
		return query, args
	}

	values := []any{}
	indexByName := map[string]int{}
	query = namedParamPattern.ReplaceAllStringFunc(query, func(match string) string {
		name := strings.TrimPrefix(match, ":")
		index, ok := indexByName[name]
		if !ok {
			index = len(values) + 1
			indexByName[name] = index
			values = append(values, namedArgs[name])
		}
		return fmt.Sprintf("$%d", index)
	})

	return query, values
}
