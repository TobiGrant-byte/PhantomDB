package store

func (s *Store) AddTagToNote(noteID, tagName string) error {
	if err := s.DB.Put([]byte("tag:"+tagName+":"+noteID), []byte{1}); err != nil {
		return err
	}
	return s.DB.Put([]byte("note_tags:"+noteID+":"+tagName), []byte{1})
}

func (s *Store) RemoveTagFromNote(noteID, tagName string) error {
	s.DB.Delete([]byte("tag:" + tagName + ":" + noteID))
	return s.DB.Delete([]byte("note_tags:" + noteID + ":" + tagName))
}

func (s *Store) ListTagsForNote(noteID string) ([]string, error) {
	pairs, err := s.DB.Scan([]byte("note_tags:"+noteID+":"), []byte("note_tags:"+noteID+":\xff"))
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, p := range pairs {
		tags = append(tags, extractLastSegment(string(p.Key)))
	}
	return tags, nil
}

func (s *Store) ListNotesByTag(tagName string) ([]string, error) {
	pairs, err := s.DB.Scan([]byte("tag:"+tagName+":"), []byte("tag:"+tagName+":\xff"))
	if err != nil {
		return nil, err
	}
	var noteIDs []string
	for _, p := range pairs {
		noteIDs = append(noteIDs, extractLastSegment(string(p.Key)))
	}
	return noteIDs, nil
}