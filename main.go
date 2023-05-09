package main

import (
	"bytes"
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

// Returns the first child node whose range intersects the key (child[i] <= key)
// TODO: bisect
func NodeLookupLE(node BNode, key []byte) uint16 {
	nKeys := node.NKeys()
	found := uint16(0)

	// the first key is a copy from the parent node,
	// therefore it's always less than or equal to the key
	for i := uint16(1); i < nKeys; i += 1 {
		cmp := bytes.Compare(node.GetKey(i), key)
		if cmp <= 0 {
			found = i
		}

		if cmp >= 0 {
			break
		}
	}

	return found
}

// Add a new key to a leaf node
func LeafInsert(newNode BNode, oldNode BNode, idx uint16, key []byte, val []byte) {
	newNode.SetHeader(BNODE_LEAF, oldNode.NKeys()+1)
	NodeAppendRange(newNode, oldNode, 0, 0, idx)
	NodeAppendKV(newNode, idx, 0, key, val)
	NodeAppendRange(newNode, oldNode, idx+1, idx, oldNode.NKeys()-idx)
}

// Copy multiple KVs into position
func NodeAppendRange(newNode BNode, oldNode BNode, dstNew uint16, srcOld uint16, n uint16) {
	if srcOld+n <= oldNode.NKeys() {
		panic("srcOld+n <= oldNode.NKeys()")
	} else if dstNew+n <= newNode.NKeys() {
		panic("dstNew+n <= newNode.NKeys()")
	} else if n == 0 {
		return
	}

	// pointers
	for i := uint16(0); i < n; i += 1 {
		newNode.SetPtr(dstNew+i, oldNode.GetPtr(srcOld+i))
	}

	// offsets
	dstBegin := newNode.GetOffset(dstNew)
	srcBegin := oldNode.GetOffset(srcOld)
	for i := uint16(1); i <= n; i += 1 {
		offset := dstBegin + oldNode.GetOffset(srcOld+i) - srcBegin
		newNode.SetOffset(dstNew+i, offset)
	}

	// KVs
	begin := oldNode.KVPosition(srcOld)
	end := oldNode.KVPosition(srcOld + n)
	copy(newNode.Data[newNode.KVPosition(dstNew):], oldNode.Data[begin:end])
}

// Copy a KV into the position
func NodeAppendKV(newNode BNode, idx uint16, ptr uint64, key []byte, val []byte) {
	// pointers
	newNode.SetPtr(idx, ptr)

	// KVs
	keyLen := len(key)
	valLen := len(val)
	position := newNode.KVPosition(idx)
	binary.LittleEndian.PutUint16(newNode.Data[position+0:], uint16(keyLen))
	binary.LittleEndian.PutUint16(newNode.Data[position+2:], uint16(valLen))
	copy(newNode.Data[position+4:], key)
	copy(newNode.Data[position+4+uint16(keyLen):], val)

	// The offset of the next key
	newNode.SetOffset(idx+1, newNode.GetOffset(idx)+4+uint16(keyLen+valLen))
}

// Part of TreeInsert(): KV insertion to an internal node
func NodeInsert(tree *BTree, newNode BNode, node BNode, idx uint16, key []byte, val []byte) {
	// Get and deallocate the child node
	cPtr := node.GetPtr(idx)
	cNode := tree.Get(cPtr)
	tree.Del(cPtr)

	// Recursive insertion to the child node
	cNode = TreeInsert(tree, cNode, key, val)

	// Split the result
	nSplit, splited := nodeSplit3(cNode)

	// Update the child links
	NodeReplaceChildN(tree, newNode, node, idx, splited[:nSplit]...)
}

// Insert a KV into a node, the result might be split into 2 nodes
// The caller is responsible for deallocating the input
// and splitting and allocating result nodes
func TreeInsert(tree *BTree, node BNode, key []byte, val []byte) BNode {
	// The result node. It's allowed to be bigger than 1 page and will be split
	// if so.
	newNode := BNode{
		Data: make([]byte, 2*BTREE_PAGE_SIZE),
	}

	// Where to insert the key?
	idx := NodeLookupLE(node, key)

	// Act depending on node type
	switch node.BType() {
	case BNODE_LEAF:
		// Leaf, node.GetKey(idx) <= key
		if bytes.Equal(key, node.GetKey(idx)) {
			// Found the key, update it
			LeafUpdate(newNode, node, idx, key, val)
		} else {
			// Insert it after the position
			LeafInsert(newNode, node, idx+1, key, val)
		}
	case BNODE_NODE:
		// Internal node, insert it to a child node
		NodeInsert(tree, newNode, node, idx, key, val)
	default:
		panic("Uknown node type")
	}

	return newNode
}

func main() {
}
