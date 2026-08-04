package phantomdb

import (
	"os"
	"testing"
)

func TestWALReplaySurvivesRestart(t *testing.T) {
	path := "test.wal"
	defer os.Remove(path)

	wal, _ := OpenWAL(path)
	wal.Append(WALRecord{Op: WALOpPut, Key: []byte("k1"), Value: []byte("v1")})
	wal.Append(WALRecord{Op: WALOpPut, Key: []byte("k2"), Value: []byte("v2")})
	wal.Close() // simulate process exit WITHOUT a checkpoint

	wal2, _ := OpenWAL(path)
	defer wal2.Close()

	var recovered []WALRecord
	err := wal2.Replay(func(rec WALRecord) error {
		recovered = append(recovered, rec)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 2 {
		t.Fatalf("expected 2 records recovered, got %d", len(recovered))
	}
	if string(recovered[0].Key) != "k1" || string(recovered[1].Key) != "k2" {
		t.Fatalf("recovered records in wrong order or content: %+v", recovered)
	}
}