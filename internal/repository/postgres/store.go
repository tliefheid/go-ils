package postgres

import (
	"database/sql"
	"os"

	"github.com/yourusername/library-ils-backend/internal/repository"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (repository.Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(fn string) error {
	data, err := os.ReadFile(fn)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(string(data))

	return err
}
