package phantomdb

import (
	"os"
	"testing"
)

func TestIndexAndSearchText(t *testing.T) {
	base := "test_index_db"
	defer os.Remove(base + ".pdb")
	defer os.Remove(base + ".wal")

	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.IndexText("note:1", "Learning Go and building a database engine"); err != nil {
		t.Fatal(err)
	}
	if err := db.IndexText("note:2", "Cooking pasta with garlic and olive oil"); err != nil {
		t.Fatal(err)
	}

	// "database" should only match note:1
	results, err := db.SearchText("database")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0] != "note:1" {
		t.Fatalf("expected [note:1], got %v", results)
	}

	// "and" appears in both notes
	results, err = db.SearchText("and")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches for 'and', got %v", results)
	}

	// word that appears nowhere
	results, err = db.SearchText("javascript")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no matches for 'javascript', got %v", results)
	}
}