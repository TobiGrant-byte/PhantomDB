package store

import "knowledgebase/internal/models"

func (s *Store) SearchNotes(query string) ([]models.Note, error) {
	docIDs, err := s.DB.SearchText(query)
	if err != nil {
		return nil, err
	}
	var notes []models.Note
	for _, docID := range docIDs {
		id := docID[len("note:"):]
		n, err := s.GetNote(id)
		if err != nil {
			continue
		}
		notes = append(notes, n)
	}
	return notes, nil
}