package phantomdb

import (
	"encoding/binary"
	"io"
	"os"
)

const (
	WALOpPut    byte = 1
	WALOpDelete byte = 2
)

type WALRecord struct {
	Op    byte
	Key   []byte
	Value []byte
}

type WAL struct {
	file *os.File
}

func OpenWAL(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: f}, nil
}

func (w *WAL) Append(rec WALRecord) error {
	buf := encodeWALRecord(rec)
	if _, err := w.file.Write(buf); err != nil {
		return err
	}
	return w.file.Sync() // MUST fsync here — this is the durability guarantee
}

func encodeWALRecord(rec WALRecord) []byte {
	buf := make([]byte, 0, 9+len(rec.Key)+len(rec.Value))
	buf = append(buf, rec.Op)
	klen := make([]byte, 4)
	binary.BigEndian.PutUint32(klen, uint32(len(rec.Key)))
	buf = append(buf, klen...)
	buf = append(buf, rec.Key...)
	vlen := make([]byte, 4)
	binary.BigEndian.PutUint32(vlen, uint32(len(rec.Value)))
	buf = append(buf, vlen...)
	buf = append(buf, rec.Value...)
	return buf
}

func readWALRecord(r io.Reader) (WALRecord, error) {
	var rec WALRecord
	opBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, opBuf); err != nil {
		return rec, err // io.EOF at a clean boundary = end of log
	}
	rec.Op = opBuf[0]

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return rec, err
	}
	klen := binary.BigEndian.Uint32(lenBuf)
	rec.Key = make([]byte, klen)
	if _, err := io.ReadFull(r, rec.Key); err != nil {
		return rec, err
	}

	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return rec, err
	}
	vlen := binary.BigEndian.Uint32(lenBuf)
	rec.Value = make([]byte, vlen)
	if _, err := io.ReadFull(r, rec.Value); err != nil {
		return rec, err
	}

	return rec, nil
}

// Replay applies every logged record via `apply`. Called once at startup,
// before the DB accepts new operations.
func (w *WAL) Replay(apply func(WALRecord) error) error {
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	for {
		rec, err := readWALRecord(w.file)
		if err != nil {
			break // EOF, or a truncated trailing record from a crash mid-write — stop safely
		}
		if err := apply(rec); err != nil {
			return err
		}
	}
	// Reset position to end for further appends
	_, err := w.file.Seek(0, io.SeekEnd)
	return err
}

func (w *WAL) Truncate() error {
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	_, err := w.file.Seek(0, io.SeekStart)
	return err
}

func (w *WAL) Close() error {
	return w.file.Close()
}