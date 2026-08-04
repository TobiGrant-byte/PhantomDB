package phantomdb

import (
	"fmt"
	"os"
	"testing"
)

func TestBTreeSingleLeafInsertSearch(t *testing.T) {
	disk, _ := OpenDiskManager("test_btree1.pdb")
	defer os.Remove("test_btree1.pdb")
	buf := NewBufferPool(disk, 16)
	tree := NewBTree(buf)

	tree.Insert([]byte("cat"), []byte("meow"))
	tree.Insert([]byte("dog"), []byte("woof"))

	v, err := tree.Search([]byte("cat"))
	if err != nil || string(v) != "meow" {
		t.Fatalf("expected meow, got %s, err %v", v, err)
	}
}

func TestBTreeSplitting(t *testing.T) {
	disk, _ := OpenDiskManager("test_btree2.pdb")
	defer os.Remove("test_btree2.pdb")
	buf := NewBufferPool(disk, 64)
	tree := NewBTree(buf)

	for i := 0; i < 300; i++ {
		key := fmt.Sprintf("key%04d", i)
		tree.Insert([]byte(key), []byte(fmt.Sprintf("val%04d", i)))
	}

	for i := 0; i < 300; i++ {
		key := fmt.Sprintf("key%04d", i)
		v, err := tree.Search([]byte(key))
		if err != nil {
			t.Fatalf("key%04d not found after split: %v", i, err)
		}
		expected := fmt.Sprintf("val%04d", i)
		if string(v) != expected {
			t.Fatalf("wrong value for %s: got %s want %s", key, v, expected)
		}
	}

	results, err := tree.Scan([]byte("key0000"), []byte("key9999"))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 300 {
		t.Fatalf("expected 300 scan results, got %d", len(results))
	}
}


func TestBTreeDelete(t *testing.T) {
	disk, _ := OpenDiskManager("test_btree_delete.pdb")
	defer os.Remove("test_btree_delete.pdb")
	buf := NewBufferPool(disk, 16)
	tree := NewBTree(buf)

	tree.Insert([]byte("cat"), []byte("meow"))
	tree.Insert([]byte("dog"), []byte("woof"))

	if err := tree.Delete([]byte("cat")); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	_, err := tree.Search([]byte("cat"))
	if err != ErrKeyNotFound {
		t.Fatalf("expected cat to be gone, got err=%v", err)
	}

	v, err := tree.Search([]byte("dog"))
	if err != nil || string(v) != "woof" {
		t.Fatalf("dog should still exist, got %s, err %v", v, err)
	}

	err = tree.Delete([]byte("nonexistent"))
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for missing key, got %v", err)
	}
}