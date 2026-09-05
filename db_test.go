package phantomdb

import (
	"os"
	"testing"
	"fmt"
	"sync"
)

func TestDBCrashRecovery(t *testing.T) {
	base := "test_crash_db"
	defer os.Remove(base + ".pdb")
	defer os.Remove(base + ".wal")

	db, _ := Open(base)
	db.Put([]byte("k1"), []byte("v1"))
	db.Put([]byte("k2"), []byte("v2"))
	// Deliberately DO NOT call Checkpoint or Close — simulates a hard crash

	db2, err := Open(base) // reopen as if the process restarted
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	v1, err := db2.Get([]byte("k1"))
	if err != nil || string(v1) != "v1" {
		t.Fatalf("k1 not recovered: %v, %v", v1, err)
	}
	v2, err := db2.Get([]byte("k2"))
	if err != nil || string(v2) != "v2" {
		t.Fatalf("k2 not recovered: %v, %v", v2, err)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	base := "test_concurrent_db"
	defer os.Remove(base + ".pdb")
	defer os.Remove(base + ".wal")

	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var writers sync.WaitGroup
	var readers sync.WaitGroup
	stopReading := make(chan struct{})

	for w := 0; w < 10; w++ {
		writers.Add(1)
		go func(writerID int) {
			defer writers.Done()
			for i := 0; i < 50; i++ {
				key := fmt.Sprintf("writer%d-key%d", writerID, i)
				val := fmt.Sprintf("val%d-%d", writerID, i)
				if err := db.Put([]byte(key), []byte(val)); err != nil {
					t.Errorf("put failed: %v", err)
				}
			}
		}(w)
	}

	for r := 0; r < 5; r++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stopReading:
					return
				default:
					db.Scan([]byte("writer"), []byte("writer\xff"))
				}
			}
		}()
	}

	writers.Wait()
	close(stopReading)
	readers.Wait()

	all, err := db.Scan([]byte("writer"), []byte("writer\xff"))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 500 {
		t.Fatalf("expected 500 keys after concurrent writes, got %d", len(all))
	}
}

func TestTransactionCommit(t *testing.T) {
	base := "test_tx_commit"
	defer os.Remove(base + ".pdb")
	defer os.Remove(base + ".wal")

	db, _ := Open(base)
	defer db.Close()

	tx := db.Begin()
	tx.Put([]byte("a"), []byte("1"))
	tx.Put([]byte("b"), []byte("2"))
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	va, _ := db.Get([]byte("a"))
	vb, _ := db.Get([]byte("b"))
	if string(va) != "1" || string(vb) != "2" {
		t.Fatalf("commit didn't apply both writes: a=%s b=%s", va, vb)
	}
}

func TestTransactionAbort(t *testing.T) {
	base := "test_tx_abort"
	defer os.Remove(base + ".pdb")
	defer os.Remove(base + ".wal")

	db, _ := Open(base)
	defer db.Close()

	tx := db.Begin()
	tx.Put([]byte("x"), []byte("should-not-appear"))
	tx.Abort()

	_, err := db.Get([]byte("x"))
	if err != ErrKeyNotFound {
		t.Fatalf("expected x to not exist after abort, got err=%v", err)
	}
}