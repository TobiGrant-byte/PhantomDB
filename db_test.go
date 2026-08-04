package phantomdb

import (
	"os"
	"testing"
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