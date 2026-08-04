package phantomdb

import "testing"

func TestPageInsertAndRead(t *testing.T) {
	p := NewPage(1, 1) // type 1 = leaf

	if !p.InsertLeafCell([]byte("banana"), []byte("yellow")) {
		t.Fatal("expected insert to succeed")
	}
	if !p.InsertLeafCell([]byte("apple"), []byte("red")) {
		t.Fatal("expected insert to succeed")
	}

	cells := p.readAllLeafCells()
	if len(cells) != 2 {
		t.Fatalf("expected 2 cells, got %d", len(cells))
	}
	// Must be sorted: apple before banana
	if string(cells[0].Key) != "apple" || string(cells[1].Key) != "banana" {
		t.Fatalf("cells not sorted: got %s, %s", cells[0].Key, cells[1].Key)
	}
}