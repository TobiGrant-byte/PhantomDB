package phantomdb

import (
	"encoding/binary"
	"errors"
)

var ErrPageFull = errors.New("page full, split required")
var ErrKeyNotFound = errors.New("key not found")

type BTree struct {
	buffer   *BufferPool
	rootPage PageID
}

type KVPair struct {
	Key   []byte
	Value []byte
}

func NewBTree(buffer *BufferPool) *BTree {
	root := buffer.NewPage(1)
	root.SetRightSibling(InvalidPageID)
	root.SetParentPage(InvalidPageID)
	return &BTree{buffer: buffer, rootPage: root.ID}
}

func (t *BTree) Search(key []byte) ([]byte, error) {
	leafID, err := t.findLeaf(key)
	if err != nil {
		return nil, err
	}
	page, err := t.buffer.FetchPage(leafID)
	if err != nil {
		return nil, err
	}
	for _, cell := range page.readAllLeafCells() {
		if compareBytes(cell.Key, key) == 0 {
			return cell.Value, nil
		}
	}
	return nil, ErrKeyNotFound
}

func (t *BTree) Insert(key, value []byte) error {
	leafID, err := t.findLeaf(key)
	if err != nil {
		return err
	}
	leaf, err := t.buffer.FetchPage(leafID)
	if err != nil {
		return err
	}
	if leaf.InsertLeafCell(key, value) {
		t.buffer.MarkDirty(leafID)
		return nil
	}
	return t.splitLeafAndInsert(leaf, key, value)
}

func (t *BTree) findLeaf(key []byte) (PageID, error) {
	id := t.rootPage
	for {
		page, err := t.buffer.FetchPage(id)
		if err != nil {
			return InvalidPageID, err
		}
		if page.Type() == 1 {
			return id, nil
		}
		child0, cells := readAllInternalCells(page)
		next := child0
		for _, c := range cells {
			if compareBytes(key, c.Key) < 0 {
				break
			}
			next = c.Child
		}
		id = next
	}
}

func (t *BTree) findParent(childID PageID) (PageID, error) {
	page, err := t.buffer.FetchPage(childID)
	if err != nil {
		return InvalidPageID, err
	}
	parent := page.ParentPage()
	if parent == InvalidPageID {
		return InvalidPageID, errors.New("no parent (this is the root)")
	}
	return parent, nil
}

func (t *BTree) splitLeafAndInsert(leaf *Page, key, value []byte) error {
	cells := leaf.readAllLeafCells()

	insertAt := len(cells)
	for i, c := range cells {
		if compareBytes(key, c.Key) < 0 {
			insertAt = i
			break
		}
	}
	newCell := leafCell{Key: key, Value: value}
	all := append(cells[:insertAt:insertAt], append([]leafCell{newCell}, cells[insertAt:]...)...)

	mid := len(all) / 2
	leftCells, rightCells := all[:mid], all[mid:]

	newRight := t.buffer.NewPage(1)
	newRight.SetRightSibling(leaf.RightSibling())
	newRight.SetParentPage(leaf.ParentPage())
	newRight.rewriteLeafCells(rightCells)

	leaf.SetRightSibling(newRight.ID)
	leaf.rewriteLeafCells(leftCells)
	t.buffer.MarkDirty(leaf.ID)
	t.buffer.MarkDirty(newRight.ID)

	return t.insertIntoParent(leaf.ID, rightCells[0].Key, newRight.ID)
}

func (t *BTree) insertIntoParent(leftID PageID, sepKey []byte, rightID PageID) error {
	if leftID == t.rootPage {
		newRoot := t.buffer.NewPage(2)
		newRoot.SetParentPage(InvalidPageID)
		writeInternalRoot(newRoot, leftID, sepKey, rightID)
		t.rootPage = newRoot.ID
		t.buffer.MarkDirty(newRoot.ID)

		left, _ := t.buffer.FetchPage(leftID)
		left.SetParentPage(newRoot.ID)
		t.buffer.MarkDirty(leftID)

		right, _ := t.buffer.FetchPage(rightID)
		right.SetParentPage(newRoot.ID)
		t.buffer.MarkDirty(rightID)
		return nil
	}

	parentID, err := t.findParent(leftID)
	if err != nil {
		return err
	}
	parent, err := t.buffer.FetchPage(parentID)
	if err != nil {
		return err
	}

	right, err := t.buffer.FetchPage(rightID)
	if err != nil {
		return err
	}
	right.SetParentPage(parentID)
	t.buffer.MarkDirty(rightID)

	if insertInternalCell(parent, sepKey, rightID) {
		t.buffer.MarkDirty(parentID)
		return nil
	}
	return t.splitInternalAndInsert(parent, sepKey, rightID)
}

func (t *BTree) splitInternalAndInsert(parent *Page, sepKey []byte, rightChildID PageID) error {
	child0, cells := readAllInternalCells(parent)

	insertAt := len(cells)
	for i, c := range cells {
		if compareBytes(sepKey, c.Key) < 0 {
			insertAt = i
			break
		}
	}
	newCell := internalCell{Key: sepKey, Child: rightChildID}
	all := append(cells[:insertAt:insertAt], append([]internalCell{newCell}, cells[insertAt:]...)...)

	mid := len(all) / 2
	midKey := all[mid].Key
	leftCells := all[:mid]
	rightCells := all[mid+1:]
	rightChild0 := all[mid].Child

	newRight := t.buffer.NewPage(2)
	newRight.SetParentPage(parent.ParentPage())
	writeInternalCells(newRight, rightChild0, rightCells)
	t.reparentChildren(newRight.ID, rightChild0, rightCells)

	writeInternalCells(parent, child0, leftCells)
	t.buffer.MarkDirty(parent.ID)
	t.buffer.MarkDirty(newRight.ID)

	return t.insertIntoParent(parent.ID, midKey, newRight.ID)
}

func (t *BTree) reparentChildren(newParentID PageID, child0 PageID, cells []internalCell) {
	if c, err := t.buffer.FetchPage(child0); err == nil {
		c.SetParentPage(newParentID)
		t.buffer.MarkDirty(child0)
	}
	for _, cell := range cells {
		if c, err := t.buffer.FetchPage(cell.Child); err == nil {
			c.SetParentPage(newParentID)
			t.buffer.MarkDirty(cell.Child)
		}
	}
}

func (t *BTree) Scan(startKey, endKey []byte) ([]KVPair, error) {
	var results []KVPair
	leafID, err := t.findLeaf(startKey)
	if err != nil {
		return nil, err
	}
	for leafID != InvalidPageID {
		leaf, err := t.buffer.FetchPage(leafID)
		if err != nil {
			return nil, err
		}
		for _, c := range leaf.readAllLeafCells() {
			if compareBytes(c.Key, endKey) > 0 {
				return results, nil
			}
			if compareBytes(c.Key, startKey) >= 0 {
				results = append(results, KVPair{Key: c.Key, Value: c.Value})
			}
		}
		leafID = leaf.RightSibling()
	}
	return results, nil
}

type internalCell struct {
	Key   []byte
	Child PageID
}

func readAllInternalCells(page *Page) (PageID, []internalCell) {
	offset := uint32(HeaderSize)
	child0 := PageID(binary.BigEndian.Uint32(page.Data[offset : offset+4]))
	offset += 4

	var cells []internalCell
	n := page.NumCells()
	for i := uint16(0); i < n; i++ {
		klen := binary.BigEndian.Uint16(page.Data[offset : offset+2])
		offset += 2
		key := append([]byte{}, page.Data[offset:offset+uint32(klen)]...)
		offset += uint32(klen)
		child := PageID(binary.BigEndian.Uint32(page.Data[offset : offset+4]))
		offset += 4
		cells = append(cells, internalCell{Key: key, Child: child})
	}
	return child0, cells
}

func writeInternalCells(page *Page, child0 PageID, cells []internalCell) {
	offset := uint32(HeaderSize)
	binary.BigEndian.PutUint32(page.Data[offset:offset+4], uint32(child0))
	offset += 4
	for _, c := range cells {
		binary.BigEndian.PutUint16(page.Data[offset:offset+2], uint16(len(c.Key)))
		offset += 2
		copy(page.Data[offset:], c.Key)
		offset += uint32(len(c.Key))
		binary.BigEndian.PutUint32(page.Data[offset:offset+4], uint32(c.Child))
		offset += 4
	}
	page.setNumCells(uint16(len(cells)))
	page.setFreeSpaceOffset(offset)
}

func writeInternalRoot(page *Page, leftID PageID, sepKey []byte, rightID PageID) {
	writeInternalCells(page, leftID, []internalCell{{Key: sepKey, Child: rightID}})
}

func insertInternalCell(page *Page, key []byte, child PageID) bool {
	need := 2 + len(key) + 4
	free := PageSize - int(page.FreeSpaceOffset())
	if need > free {
		return false
	}
	child0, cells := readAllInternalCells(page)
	insertAt := len(cells)
	for i, c := range cells {
		if compareBytes(key, c.Key) < 0 {
			insertAt = i
			break
		}
	}
	newCell := internalCell{Key: append([]byte{}, key...), Child: child}
	cells = append(cells[:insertAt:insertAt], append([]internalCell{newCell}, cells[insertAt:]...)...)
	writeInternalCells(page, child0, cells)
	return true
}


func (t *BTree) Delete(key []byte) error {
	leafID, err := t.findLeaf(key)
	if err != nil {
		return err
	}
	leaf, err := t.buffer.FetchPage(leafID)
	if err != nil {
		return err
	}

	cells := leaf.readAllLeafCells()
	found := false
	remaining := make([]leafCell, 0, len(cells))
	for _, c := range cells {
		if compareBytes(c.Key, key) == 0 {
			found = true
			continue // skip it — this is the removal
		}
		remaining = append(remaining, c)
	}

	if !found {
		return ErrKeyNotFound
	}

	leaf.rewriteLeafCells(remaining)
	t.buffer.MarkDirty(leafID)
	return nil
}