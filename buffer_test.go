package phantomdb

import (
	"os"
	"testing"
)

func TestBufferPoolFetchAndDirty(t *testing.T) {
	path := "test_buffer.pdb"
	defer os.Remove(path)

	disk, err := OpenDiskManager(path)
	if err != nil {
		t.Fatal(err)
	}
	defer disk.Close()

	buf := NewBufferPool(disk, 16)

	// Create a new page through the buffer pool and write a cell to it
	p := buf.NewPage(1) // leaf type
	p.InsertLeafCell([]byte("k1"), []byte("v1"))
	buf.MarkDirty(p.ID)

	// Fetch it again — should return the cached (dirty) version, same data
	fetched, err := buf.FetchPage(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	cells := fetched.readAllLeafCells()
	if len(cells) != 1 || string(cells[0].Key) != "k1" {
		t.Fatalf("expected cached page with k1, got %+v", cells)
	}

	// Flush to disk, then confirm it can be read back independently
	if err := buf.FlushAll(); err != nil {
		t.Fatal(err)
	}

	readBack, err := disk.ReadPage(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	cells2 := readBack.readAllLeafCells()
	if len(cells2) != 1 || string(cells2[0].Key) != "k1" {
		t.Fatalf("flushed page missing data: %+v", cells2)
	}
}