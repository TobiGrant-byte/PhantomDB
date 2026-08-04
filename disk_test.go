package phantomdb

import (
	"os"
	"testing"
)

func TestDiskManagerReadWrite(t *testing.T) {
	path := "test_disk.pdb"
	defer os.Remove(path)

	dm, err := OpenDiskManager(path)
	if err != nil {
		t.Fatal(err)
	}
	defer dm.Close()

	id := dm.AllocatePage()
	p := NewPage(id, 1)
	p.InsertLeafCell([]byte("k1"), []byte("v1"))

	if err := dm.WritePage(p); err != nil {
		t.Fatal(err)
	}
	if err := dm.Sync(); err != nil {
		t.Fatal(err)
	}

	// Simulate reopening the file fresh
	dm.Close()
	dm2, err := OpenDiskManager(path)
	if err != nil {
		t.Fatal(err)
	}
	defer dm2.Close()

	readBack, err := dm2.ReadPage(id)
	if err != nil {
		t.Fatal(err)
	}
	cells := readBack.readAllLeafCells()
	if len(cells) != 1 || string(cells[0].Key) != "k1" {
		t.Fatalf("data did not survive close/reopen: %+v", cells)
	}
}