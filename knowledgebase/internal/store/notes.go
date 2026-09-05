package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"

	"knowledgebase/internal/models"
)

func encodeNote(n models.Note) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(n)
	return buf.Bytes(), err
}

func decodeNote(data []byte) (models.Note, error) {
	var n models.Note
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&n)
	return n, err
}

func (s *Store) CreateNote(title, content string) (models.Note, error) {
	id, err := s.nextID("counter:notes")
	if err != nil {
		return models.Note{}, err
	}

	note := models.Note{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := encodeNote(note)
	if err != nil {
		return models.Note{}, err
	}
	if err := s.DB.Put([]byte("note:"+id), data); err != nil {
		return models.Note{}, err
	}

	if err := s.DB.IndexText("note:"+id, title+" "+content); err != nil {
		return models.Note{}, err
	}

	return note, nil
}

func (s *Store) GetNote(id string) (models.Note, error) {
	data, err := s.DB.Get([]byte("note:" + id))
	if err != nil {
		return models.Note{}, errors.New("note not found")
	}
	return decodeNote(data)
}

func (s *Store) ListNotes() ([]models.Note, error) {
	pairs, err := s.DB.Scan([]byte("note:"), []byte("note:\xff"))
	if err != nil {
		return nil, err
	}
	var notes []models.Note
	for _, p := range pairs {
		n, err := decodeNote(p.Value)
		if err != nil {
			continue
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *Store) UpdateNote(id, title, content string) (models.Note, error) {
	existing, err := s.GetNote(id)
	if err != nil {
		return models.Note{}, err
	}
	existing.Title = title
	existing.Content = content
	existing.UpdatedAt = time.Now()

	data, err := encodeNote(existing)
	if err != nil {
		return models.Note{}, err
	}
	if err := s.DB.Put([]byte("note:"+id), data); err != nil {
		return models.Note{}, err
	}

	if err := s.DB.IndexText("note:"+id, title+" "+content); err != nil {
		return models.Note{}, err
	}

	return existing, nil
}

// DeleteNote uses a transaction — the note, its tag memberships, and its
// backlinks are removed as one atomic group.
func (s *Store) DeleteNote(id string) error {
	tx := s.DB.Begin()

	tx.Delete([]byte("note:" + id))

	tagPairs, _ := s.DB.Scan([]byte("note_tags:"+id+":"), []byte("note_tags:"+id+":\xff"))
	for _, p := range tagPairs {
		tagName := extractLastSegment(string(p.Key))
		tx.Delete([]byte("tag:" + tagName + ":" + id))
		tx.Delete(p.Key)
	}

	blPairs, _ := s.DB.Scan([]byte("backlink:"+id+":"), []byte("backlink:"+id+":\xff"))
	for _, p := range blPairs {
		tx.Delete(p.Key)
	}

	return tx.Commit()
}

func extractLastSegment(key string) string {
	parts := splitColon(key)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func splitColon(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}