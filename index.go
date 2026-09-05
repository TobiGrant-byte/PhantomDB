package phantomdb

import "strings"

func (db *DB) IndexText(docID string, text string) error {
	for _, word := range tokenize(text) {
		key := []byte("idx:" + word + ":" + docID)
		if err := db.Put(key, []byte{1}); err != nil {
			return err
		}
	}
	return nil
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	seen := make(map[string]bool)
	var out []string
	for _, w := range fields {
		if len(w) < 3 || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

func (db *DB) SearchText(prefix string) ([]string, error) {
	prefix = strings.ToLower(prefix)
	if prefix == "" {
		return nil, nil
	}

	pairs, err := db.Scan([]byte("idx:"+prefix), []byte("idx:"+prefix+"\xff"))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var docIDs []string
	for _, p := range pairs {
		parts := strings.SplitN(string(p.Key), ":", 3)
		if len(parts) == 3 && !seen[parts[2]] {
			seen[parts[2]] = true
			docIDs = append(docIDs, parts[2])
		}
	}
	return docIDs, nil
}