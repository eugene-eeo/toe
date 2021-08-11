package eval

import (
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"math"
)

// This file implements a hash table for values.
// Values which are hashable can be inserted into the table;
// hashable objects are either one of the hashable builtins,
// or an object that implements the hash protocol.

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

var (
	ht_NIL_HASH   = htHashConstant("nil")
	ht_TRUE_HASH  = htHashConstant("true")
	ht_FALSE_HASH = htHashConstant("false")
)

// htTruncateHash truncates the given uint64 into a number that
// fits exactly into Number (float64). Golang's float64 is an
// IEEE-754 double-precision format, so it can only store up to
// 53 bits of integer precision:
//
//   https://en.wikipedia.org/wiki/Double-precision_floating-point_format#Precision_limitations_on_integer_values
//
func htTruncateHash(u uint64) Number {
	var mask uint64 = (1 << 53) - 1
	return Number(u & mask)
}

func htHashConstant(c string) Value {
	h := fnv.New64a()
	h.Write([]byte(c))
	return htTruncateHash(h.Sum64())
}

func (v Nil) Hash() Value { return ht_NIL_HASH }
func (v Boolean) Hash() Value {
	if v {
		return ht_TRUE_HASH
	} else {
		return ht_FALSE_HASH
	}
}

func (v String) Hash() Value {
	h := fnv.New64a()
	h.Write([]byte("S"))
	h.Write([]byte(v))
	return htTruncateHash(h.Sum64())
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
	return htTruncateHash(h.Sum64())
}

func (v *Function) Hash() Value {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("F%p", v)))
	return htTruncateHash(h.Sum64())
}

// =================
// Actual hash table
// =================
//
// The actual hash table is uses linear-probing. This is chosen primarily
// because it is easy to implement -- performance be damned.

const (
	ht_SIZE_HI  = 0.75 // When should we upsize?
	ht_SIZE_LO  = 0.10 // When should we downsize?
	ht_MIN_SIZE = 16
)

type htEntry struct {
	hash  uint64
	key   *Value
	value *Value
}

func (he htEntry) hasValue() bool    { return he.key != nil }
func (he htEntry) isTombstone() bool { return he.key == nil && he.value == &TOMBSTONE }
func (he htEntry) isEmpty() bool     { return he.key == nil && he.value == nil }

type hashTable struct {
	ctx     *Context
	entries []htEntry
	sz      uint64 // number of non-free entries in the hash table
	realSz  uint64 // number of non-tombstone, and non-empty entries in the hash table
}

func newHashTable(ctx *Context) *hashTable {
	return &hashTable{
		ctx:     ctx,
		entries: make([]htEntry, ht_MIN_SIZE),
		sz:      0,
	}
}

// hash tries to hash the given object -- if its hash method
// returned an error, then err will be an error object.
func (ht *hashTable) hash(k Value) (h uint64, err Value) {
	rv := ht.ctx.getObjectHash(k)
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

func (ht *hashTable) maybeResize() {
	entries := float64(len(ht.entries))
	if float64(ht.sz)/entries >= ht_SIZE_HI {
		ht.resize(true)
	} else if entries > ht_MIN_SIZE && float64(ht.realSz)/entries <= ht_SIZE_LO {
		ht.resize(false)
	}
}

func (ht *hashTable) resize(grow bool) {
	newSize := len(ht.entries)
	if grow {
		newSize *= 2
	} else {
		newSize /= 2
	}
	oldEntries := ht.entries
	ht.sz = 0
	ht.realSz = 0
	ht.entries = make([]htEntry, newSize)
	mask := uint64(len(ht.entries) - 1)
	seed := ht.ctx.ht_seed
	for _, he := range oldEntries {
		if !he.hasValue() {
			continue
		}
		// fast reinsert using .hash
		idx := (he.hash ^ seed) & mask
		start := idx
		for {
			// Note: here we only have to care about whether a key
			// was already set.
			if ht.entries[idx].key == nil {
				ht.sz++
				ht.realSz++
				ht.entries[idx] = he
				break
			}
			idx = (idx + 1) & mask
			if idx == start {
				panic("wtf!")
			}
		}
	}
}

// getEntry returns the htEntry (NOT the value) associated with
// k in the hash table, if any. If forInsert is set to true, then
// any tombstones in the chain will be returned; otherwise getEntry
// will _never_ return a tombstone. The possible return values are:
//
//   - entry != nil && err == nil   (empty entry / a matching entry)
//   - entry == nil && err != nil   (error)
func (ht *hashTable) getEntry(k Value, forInsert bool) (entry *htEntry, hash uint64, err Value) {
	hash, err = ht.hash(k)
	if err != nil {
		return nil, hash, err
	}
	size := uint64(len(ht.entries))
	mask := size - 1
	idx := (hash ^ ht.ctx.ht_seed) & mask
	start := idx
	for {
		ref := &ht.entries[idx]
		if ref.isTombstone() {
			// tombstone ==> continue probing, unless we need to insert
			if forInsert {
				return ref, hash, nil
			}
		} else if ref.isEmpty() {
			// empty entry ==> we can break the search chain
			return ref, hash, nil
		} else if ref.hash == hash {
			// potential match ==> go through the motions...
			key := *ref.key
			if key == k {
				return ref, hash, nil
			}
			cmp_res := ht.ctx.areObjectsEqual(key, k)
			if isError(cmp_res) {
				return nil, hash, cmp_res
			}
			if isTruthy(cmp_res) {
				return ref, hash, nil
			}
		}
		idx = (idx + 1) & mask
		if idx == start {
			// table is full
			return nil, hash, nil
		}
	}
}

// get finds the value associated with the given key in the hash table, if any.
func (ht *hashTable) get(k Value) (v Value, found bool, err Value) {
	entry, _, err := ht.getEntry(k, false)
	if err != nil {
		return nil, false, err
	}
	if entry == nil || entry.isEmpty() {
		return nil, false, nil
	}
	return *entry.value, true, nil
}

// delete deletes the given key from the table, if it exists.
func (ht *hashTable) delete(k Value) (found bool, err Value) {
	entry, _, err := ht.getEntry(k, false)
	if err != nil {
		return false, err
	}
	if entry == nil || entry.isEmpty() {
		return false, nil
	}
	entry.key = nil
	entry.value = &TOMBSTONE
	ht.realSz--
	ht.maybeResize()
	return true, nil
}

// insert inserts the given pair into the hash table.
// on a successful insert, err == nil.
func (ht *hashTable) insert(k Value, v Value) (err Value) {
	entry, hash, err := ht.getEntry(k, true)
	if err != nil {
		return err
	}
	if entry == nil {
		ht.resize(true)
		return ht.insert(k, v)
	}
	if entry.key == nil {
		ht.realSz++
		if entry.value != &TOMBSTONE {
			ht.sz++
		}
	}
	entry.hash = hash
	entry.key = &k
	entry.value = &v
	ht.maybeResize()
	return nil
}

func (ht *hashTable) size() uint64 {
	return ht.realSz
}

// ===============
// Utility Methods
// ===============

// getObjectHash calls the object's hash function on v and returns the result.
// if the result is an error, this needs to be handled.
func (ctx *Context) getObjectHash(v Value) (hash Value) {
	if hashable, ok := v.(interface{ Hash() Value }); ok {
		return hashable.Hash()
	}
	return newError(String(fmt.Sprintf("object %s is not hashable", v.Type())))
}
