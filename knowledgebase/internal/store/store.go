package store

import (
	"encoding/binary"
	"strconv"

	"github.com/TobiGrant-byte/phantomdb"
)

type Store struct {
	DB *phantomdb.DB
}

func Open(path string) (*Store, error) {
	db, err := phantomdb.Open(path)
	if err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) nextID(counterKey string) (string, error) {
	var n uint64
	data, err := s.DB.Get([]byte(counterKey))
	if err == nil {
		n = binary.BigEndian.Uint64(data)
	}
	n++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	if err := s.DB.Put([]byte(counterKey), buf); err != nil {
		return "", err
	}
	return strconv.FormatUint(n, 10), nil
}