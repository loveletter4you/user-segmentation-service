package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type Storage struct {
	db                *sql.DB
	userRepository    *UserRepository
	segmentRepository *SegmentRepository
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Open(dbHost, dbPort, dbRoot, dbPassword, dbName string,
	connectionAttempt, connectionTimeout int) error {
	var (
		db  *sql.DB
		err error
	)
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbRoot, dbPassword, dbHost, dbPort, dbName)
	for connectionAttempt > 0 {
		db, err = sql.Open("postgres", dbUrl)
		if err != nil {
			break
		}

		if err = db.Ping(); err == nil {
			s.db = db
			break
		}
		time.Sleep(time.Duration(connectionTimeout) * time.Second)
		connectionAttempt--
	}

	return err
}

func (s *Storage) Close() error {
	err := s.db.Close()
	return err
}

func (s *Storage) StartTransaction() (*sql.Tx, error) {
	tx, err := s.db.Begin()
	return tx, err
}

func (s *Storage) DoQuery(tx *sql.Tx, query string) (*sql.Rows, error) {
	if tx != nil {
		return tx.Query(query)
	}
	return s.db.Query(query)
}

func (s *Storage) DoQueryRow(tx *sql.Tx, query string) *sql.Row {
	if tx != nil {
		return tx.QueryRow(query)
	}
	return s.db.QueryRow(query)
}

func (s *Storage) User() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		storage: s,
	}

	return s.userRepository
}

func (s *Storage) Segment() *SegmentRepository {
	if s.segmentRepository != nil {
		return s.segmentRepository
	}

	s.segmentRepository = &SegmentRepository{
		storage: s,
	}

	return s.segmentRepository
}
