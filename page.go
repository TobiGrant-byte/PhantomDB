package phantomdb

import "encoding/binary"

type Page struct {
	ID   PageID
	Data [PageSize]byte
}

func NewPage(id PageID, pageType byte) *Page {
	p := &Page{ID: id}
	p.Data[0] = pageType
	binary.BigEndian.PutUint32(p.Data[3:7], HeaderSize) // FreeSpaceOffset starts right after header
	binary.BigEndian.PutUint32(p.Data[7:11], uint32(InvalidPageID))
	binary.BigEndian.PutUint32(p.Data[11:15], uint32(InvalidPageID))
	return p
}

func (p *Page) Type() byte             { return p.Data[0] }
func (p *Page) NumCells() uint16       { return binary.BigEndian.Uint16(p.Data[1:3]) }
func (p *Page) FreeSpaceOffset() uint32 { return binary.BigEndian.Uint32(p.Data[3:7]) }
func (p *Page) RightSibling() PageID   { return PageID(binary.BigEndian.Uint32(p.Data[7:11])) }

func (p *Page) setNumCells(n uint16) {
	binary.BigEndian.PutUint16(p.Data[1:3], n)
}
func (p *Page) setFreeSpaceOffset(off uint32) {
	binary.BigEndian.PutUint32(p.Data[3:7], off)
}
func (p *Page) SetRightSibling(id PageID) {
	binary.BigEndian.PutUint32(p.Data[7:11], uint32(id))
}

// InsertLeafCell appends a key-value cell and keeps cells sorted by key.
// Returns false if there isn't enough free space.
func (p *Page) InsertLeafCell(key, value []byte) bool {
	need := 2 + len(key) + 2 + len(value)
	free := PageSize - int(p.FreeSpaceOffset())
	if need > free {
		return false
	}

	// Find sorted insert position among existing cells
	cells := p.readAllLeafCells()
	insertAt := len(cells)
	for i, c := range cells {
		if compareBytes(key, c.Key) < 0 {
			insertAt = i
			break
		}
		if compareBytes(key, c.Key) == 0 {
			cells[i].Value = value // overwrite existing key
			p.rewriteLeafCells(cells)
			return true
		}
	}

	newCell := leafCell{Key: append([]byte{}, key...), Value: append([]byte{}, value...)}
	cells = append(cells[:insertAt], append([]leafCell{newCell}, cells[insertAt:]...)...)
	p.rewriteLeafCells(cells)
	return true
}

type leafCell struct {
	Key   []byte
	Value []byte
}

func (p *Page) readAllLeafCells() []leafCell {
	var cells []leafCell
	offset := uint32(HeaderSize)
	n := p.NumCells()
	for i := uint16(0); i < n; i++ {
		klen := binary.BigEndian.Uint16(p.Data[offset : offset+2])
		offset += 2
		key := p.Data[offset : offset+uint32(klen)]
		offset += uint32(klen)
		vlen := binary.BigEndian.Uint16(p.Data[offset : offset+2])
		offset += 2
		val := p.Data[offset : offset+uint32(vlen)]
		offset += uint32(vlen)
		cells = append(cells, leafCell{Key: append([]byte{}, key...), Value: append([]byte{}, val...)})
	}
	return cells
}

// rewriteLeafCells clears the page body and rewrites all cells in order.
// Simple and correct; not the fastest approach, but fine for this scale.
func (p *Page) rewriteLeafCells(cells []leafCell) {
	offset := uint32(HeaderSize)
	for _, c := range cells {
		binary.BigEndian.PutUint16(p.Data[offset:offset+2], uint16(len(c.Key)))
		offset += 2
		copy(p.Data[offset:], c.Key)
		offset += uint32(len(c.Key))
		binary.BigEndian.PutUint16(p.Data[offset:offset+2], uint16(len(c.Value)))
		offset += 2
		copy(p.Data[offset:], c.Value)
		offset += uint32(len(c.Value))
	}
	p.setNumCells(uint16(len(cells)))
	p.setFreeSpaceOffset(offset)
}

func compareBytes(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return int(a[i]) - int(b[i])
		}
	}
	return len(a) - len(b)
}

func (p *Page) ParentPage() PageID {
	return PageID(binary.BigEndian.Uint32(p.Data[11:15]))
}
func (p *Page) SetParentPage(id PageID) {
	binary.BigEndian.PutUint32(p.Data[11:15], uint32(id))
}