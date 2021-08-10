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
	h.Write([]byte("S"))
	h.Write([]byte(v))
	return Number(math.Float64frombits(h.Sum64()))
}

func (v Number) Hash() Value {
	h := fnv.New64a()
	floatbits := math.Float64bits(float64(v))
	var b [8]byte
	b[0] = (byte(floatbits & 0xFF))
	b[1] = (byte(floatbits >> 8 & 0xFF))
	b[2] = (byte(floatbits >> 16 & 0xFF))
	b[3] = (byte(floatbits >> 24 & 0xFF))
	b[4] = (byte(floatbits >> 32 & 0xFF))
	b[5] = (byte(floatbits >> 40 & 0xFF))
	b[6] = (byte(floatbits >> 48 & 0xFF))
	b[7] = (byte(floatbits >> 56 & 0xFF))
	h.Write([]byte("N"))
	h.Write(b[:])
	return Number(math.Float64frombits(h.Sum64()))
}

// =================
// Actual hash table
// =================

const (
	ht_INSERT_RESIZE_THRESHOLD = 0.75
	ht_INITIAL_SIZE            = 16
)

type htEntry struct {
	hash  uint64
	key   *Hashable
	value *Value
}

func (he htEntry) isTombstone() bool { return he.key == nil && he.value == &TOMBSTONE }
func (he htEntry) isEmpty() bool     { return he.key == nil && he.value == nil }

type hashTable struct {
	ctx     *Context
	entries []htEntry
	seed    uint64 // seed
	sz      uint64 // number of non-free entries in the hash table
	realSz  uint64 // number of non-tombstone, and non-empty entries in the hash table
}

func newHashTable(ctx *Context) *hashTable {
	return &hashTable{
		ctx:     ctx,
		entries: make([]htEntry, ht_INITIAL_SIZE),
		seed:    getNewHashTableSeed(),
		sz:      0,
	}
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
	return h, nil
}

func getNewHashTableSeed() uint64 {
	var b [8]byte
	rand.Read(b[:])
	rv := uint64(b[0])
	rv = (rv << 8) + uint64(b[1])
	rv = (rv << 8) + uint64(b[2])
	rv = (rv << 8) + uint64(b[3])
	rv = (rv << 8) + uint64(b[4])
	rv = (rv << 8) + uint64(b[5])
	rv = (rv << 8) + uint64(b[6])
	rv = (rv << 8) + uint64(b[7])
	return rv
}

func (ht *hashTable) resize() {
	oldEntries := ht.entries
	ht.sz = 0
	ht.realSz = 0
	ht.entries = make([]htEntry, len(ht.entries)*2)
	ht.seed = getNewHashTableSeed()
	for _, he := range oldEntries {
		if he.key != nil {
			ht.insert(*he.key, *he.value)
		}
	}
}

// getEntry returns the htEntry (NOT the value) associated with
// k in the hash table, if any. The possible return values are:
//
//   - entry != nil && err == nil   (empty entry / a matching entry)
//   - entry == nil && err != nil   (error)
func (ht *hashTable) getEntry(k Hashable) (entry *htEntry, hash uint64, err Value) {
	hash, err = ht.hash(k)
	if err != nil {
		return nil, hash, err
	}
	size := uint64(len(ht.entries))
	mask := size - 1
	seeded_hash := hash ^ ht.seed
	for i := uint64(0); i < size; i++ {
		ref := &ht.entries[(seeded_hash+i)&mask]
		if ref.isTombstone() {
			// have to keep probing.
			continue
		}
		if ref.isEmpty() {
			return ref, hash, nil
		}
		// first check for a match on the hash...
		if ref.hash == hash {
			key := *ref.key
			// fast case
			if key == k {
				return ref, hash, nil
			}
			cmp_res := ht.ctx.binary(lexer.EQUAL_EQUAL, key.(Value), k.(Value))
			if isError(cmp_res) {
				return nil, hash, cmp_res
			}
			if isTruthy(cmp_res) {
				return ref, hash, nil
			}
		}
	}
	return nil, hash, nil
}

// get finds the value associated with the given key in the hash table,
// if any. the possible return types are:
//   1. v == nil && !found   (not found)
//   2. v != nil && !found   (error)
//   3. v != nil && found    (found)
func (ht *hashTable) get(k Hashable) (v Value, found bool) {
	entry, _, err := ht.getEntry(k)
	if err != nil {
		return err, false
	}
	if entry == nil || entry.isEmpty() {
		return nil, false
	}
	return *entry.value, true
}

// delete deletes the given key from the table, if it exists.
func (ht *hashTable) delete(k Hashable) (found bool, err Value) {
	entry, _, err := ht.getEntry(k)
	if err != nil {
		return false, err
	}
	if entry == nil || entry.isEmpty() {
		return false, nil
	}
	entry.key = nil
	entry.value = &TOMBSTONE
	ht.realSz--
	return true, nil
}

func (ht *hashTable) insert(k Hashable, v Value) (err Value) {
	entry, hash, err := ht.getEntry(k)
	if err != nil {
		return err
	}
	if entry == nil {
		ht.resize()
		return ht.insert(k, v)
	}
	ht.realSz++
	if entry.isEmpty() {
		ht.sz++
	}
	entry.hash = hash
	entry.key = &k
	entry.value = &v
	if float64(ht.sz)/float64(len(ht.entries)) >= ht_INSERT_RESIZE_THRESHOLD {
		ht.resize()
	}
	return nil
}

func (ht *hashTable) size() uint64 {
	return ht.realSz
}
