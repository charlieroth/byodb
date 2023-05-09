package main

import (
	"encoding/binary"
)

const (
	BNODE_NODE = 1 // internal nodes without values
	BNODE_LEAF = 2 // leaf nodes with values
)

const HEADER = 4
const BTREE_PAGE_SIZE = 4096
const BTREE_MAX_KEY_SIZE = 1000
const BTREE_MAX_VAL_SIZE = 3000

type BNode struct {
	Data []byte // can be dumped to disk
}

// Header Functions
// See Section04#"B-Tree Node Format"
func (node BNode) BType() uint16 {
	// note: Uint16 only reads the first two bytes of a slice
	return binary.LittleEndian.Uint16(node.Data)
}

func (node BNode) NKeys() uint16 {
	return binary.LittleEndian.Uint16(node.Data[2:4])
}

func (node BNode) SetHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node.Data[0:2], btype)
	binary.LittleEndian.PutUint16(node.Data[2:4], nkeys)
}

// Pointer functions
func (node BNode) GetPtr(idx uint16) uint64 {
	if idx < node.NKeys() {
		pos := HEADER + 8*idx
		return binary.LittleEndian.Uint64(node.Data[pos:])
	}
	panic("idx is >= node.NKeys()")
}

func (node BNode) SetPtr(idx uint16, val uint64) {
	if idx < node.NKeys() {
		pos := HEADER + 8*idx
		binary.LittleEndian.PutUint64(node.Data[pos:], val)
		return
	}
	panic("idx is >= node.NKeys()")
}

// Offset List Functions
// See Section04#"Offset List Details"
func OffsetPosition(node BNode, idx uint16) uint16 {
	if idx >= 1 && idx <= node.NKeys() {
		return HEADER + 8*node.NKeys() + 2*(idx-1)
	}
	panic("idx not between 1 and node.NKeys()")
}

func (node BNode) GetOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}

	offsetPosition := OffsetPosition(node, idx)
	return binary.LittleEndian.Uint16(node.Data[offsetPosition:])
}

func (node BNode) SetOffset(idx uint16, offset uint16) {
	offsetPosition := OffsetPosition(node, idx)
	binary.LittleEndian.PutUint16(node.Data[offsetPosition:], offset)
}

// Key-Value Functions
func (node BNode) KVPosition(idx uint16) uint16 {
	if idx <= node.NKeys() {
		return HEADER + 8*node.NKeys() + 2*node.NKeys() + node.GetOffset(idx)
	}
	panic("idx > node.NKeys()")
}

func (node BNode) GetKey(idx uint16) []byte {
	if idx < node.NKeys() {
		pos := node.KVPosition(idx)
		klen := binary.LittleEndian.Uint16(node.Data[pos:])
		return node.Data[pos+4:][:klen]
	}
	panic("idx < node.NKeys()")
}

func (node BNode) NBytes() uint16 {
	return node.KVPosition(node.NKeys())
}

// We can't use in-memory pointers, the points are instead 64-bit integers
// referencing disk pages instead of in-memory nodes
type BTree struct {
	Root uint64 // pointer (a nonzero page number)

	// callbacks for managing on-disk pages
	Get func(uint64) BNode // dereference a pointer
	New func(BNode) uint64 // allocate a new page
	Del func(uint64)       // deallocate a page
}

func main() {
}
