package eval

import (
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"math"
	"toe/lexer"
)

// This file implements a hash table for values.
// Values which are hashable can be inserted into the table;
// hashable objects are either one of the hashable builtins,
// or an object that implements the hash protocol.

// =================
// Hashable Protocol
// =================

type Hashable interface {
	// We actually need equals() here, but we use binary()
	// for now.
	Hash() Value // this should return Number
}

func (v String) Hash() Value {
	h := fnv.New64a()
	h.Write([]byte(v))
	return Number(math.Float64frombits(h.Sum64()))
}

// so we can store tombstones
func (v tombstone) Hash() Value { return NIL }

// =================
// Actual hash table
// =================

type htEntry struct {
	key   Hashable
	hash  uint64
	value Value
}

func (he htEntry) isTombstone() bool { return he.key == TOMBSTONE }
func (he htEntry) isEmpty() bool     { return he.key == nil }
func (he htEntry) isFree() bool      { return he.isEmpty() || he.isTombstone() }

type hashTable struct {
	ctx     *Context
	entries []htEntry
	seed    uint64 // seed
	sz      uint64 // number of non-free entries in the hash table
}

// hash tries to hash the given object -- if its hash method
// returned an error, then err will be an error object.
func (ht *hashTable) hash(k Hashable) (h uint64, err Value) {
	rv := k.Hash()
	if isError(rv) {
		return 0, rv
	}
	if rv.Type() != VT_NUMBER {
		return 0, newError(String(fmt.Sprintf(
			"expected hash to return a number, got: %s",
			rv.Type())))
	}
	h = math.Float64bits(float64(rv.(Number)))
	h ^= ht.seed
	h &= ht.sz - 1
	return h, nil
}

func getNewHashTableSeed() uint64 {
	var b [8]byte
	rand.Read(b[:])
	rv := uint64(b[0])
	rv = rv<<8 + uint64(b[1])
	rv = rv<<8 + uint64(b[2])
	rv = rv<<8 + uint64(b[3])
	rv = rv<<8 + uint64(b[4])
	rv = rv<<8 + uint64(b[5])
	rv = rv<<8 + uint64(b[6])
	rv = rv<<8 + uint64(b[7])
	return rv
}

func (ht *hashTable) resize() {
	oldEntries := ht.entries
	ht.sz *= 2
	ht.entries = make([]htEntry, ht.sz)
	ht.seed = getNewHashTableSeed()
	for _, entry := range oldEntries {
		ht.insert(entry.key, entry.value)
	}
}

// getEntry returns the htEntry (NOT the value) associated with
// k in the hash table, if any. The possible return values are:
//   - entry != nil                 (in table)
//   - entry == nil && err == nil   (not in table)
//   - entry == nil && err != nil   (error)
func (ht *hashTable) getEntry(k Hashable) (entry *htEntry, err Value) {
	mask := uint64(len(ht.entries) - 1)
	hash, err := ht.hash(k)
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < ht.sz; i++ {
		j := (hash + i) & mask
		ref := &ht.entries[j]
		if ref.isEmpty() {
			break
		}
		if ref.isTombstone() {
			continue
		}
		// first check for a match on the hash...
		if ref.hash == hash {
			cmp_res := ht.ctx.binary(lexer.EQUAL_EQUAL, ref.key.(Value), k.(Value))
			if isError(cmp_res) {
				return nil, cmp_res
			}
			if isTruthy(cmp_res) {
				return ref, nil
			}
		}
	}
	return nil, nil
}

// get finds the value associated with the given key in the hash table,
// if any. the possible return types are:
//   1. v == nil   (not found)
//   2. v != nil   (found)
//   3. isError(v) (error)
func (ht *hashTable) get(k Hashable) (v Value) {
	entry, err := ht.getEntry(k)
	if err != nil {
		return err
	}
	if entry == nil {
		return nil
	}
	return entry.value
}

func (ht *hashTable) insert(k Hashable, v Value) (err Value) {
	// calculate hash
	mask := uint64(len(ht.entries) - 1)
	hash, err := ht.hash(k)
	if err != nil {
		return err
	}
	// now start the probe sequence.
	for i := uint64(0); i < ht.sz; i++ {
		j := (hash + i) & mask
		if ht.entries[j].isFree() {
			ref := &ht.entries[j]
			if ref.isEmpty() {
				ht.sz++
			}
			ref.key = k
			ref.hash = hash
			ref.value = v
			return nil
		}
	}
	// resize, and re-insert.
	ht.resize()
	return ht.insert(k, v)
}
