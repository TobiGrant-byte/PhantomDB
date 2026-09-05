package store

import (
	"bytes"
	"encoding/gob"
	"time"

	"knowledgebase/internal/models"
)

func (s *Store) CreateBookmark(url, title, description string) (models.Bookmark, error) {
	id, err := s.nextID("counter:bookmarks")
	if err != nil {
		return models.Bookmark{}, err
	}
	b := models.Bookmark{
		ID: id, URL: url, Title: title, Description: description,
		CreatedAt: time.Now(),
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(b); err != nil {
		return models.Bookmark{}, err
	}
	if err := s.DB.Put([]byte("bookmark:"+id), buf.Bytes()); err != nil {
		return models.Bookmark{}, err
	}
	return b, nil
}

func (s *Store) ListBookmarks() ([]models.Bookmark, error) {
	pairs, err := s.DB.Scan([]byte("bookmark:"), []byte("bookmark:\xff"))
	if err != nil {
		return nil, err
	}
	var bookmarks []models.Bookmark
	for _, p := range pairs {
		var b models.Bookmark
		if err := gob.NewDecoder(bytes.NewReader(p.Value)).Decode(&b); err == nil {
			bookmarks = append(bookmarks, b)
		}
	}
	return bookmarks, nil
}