package eval

import (
	"fmt"
	"testing"
)

func TestHashTable(t *testing.T) {
	ctx := NewContext()
	ht := newHashTable(ctx)
	for i := 0; i < 100; i++ {
		k := String(fmt.Sprintf("key:%d", i))
		v := String(fmt.Sprintf("value:%d", i))

		mustInsert(t, ht, k, v)
		u := mustGet(t, ht, k)
		if u != v {
			t.Fatalf("expected=%#+v, got=%#+v", v, u)
		}
		if ht.size() != uint64(i+1) {
			t.Fatalf("expected ht.size()=%d, got=%d", i+1, ht.size())
		}

		// inserting the same key multiple times does not cause size to increase.
		mustInsert(t, ht, k, v)
		mustInsert(t, ht, k, v)
		mustInsert(t, ht, k, v)
		mustInsert(t, ht, k, v)
		if ht.size() != uint64(i+1) {
			t.Fatalf("expected ht.size()=%d, got=%d", i+1, ht.size())
		}

		// insert a new value under the same key -- get should return the new value.
		mustInsert(t, ht, k, NIL)
		u = mustGet(t, ht, k)
		if u != NIL {
			t.Fatalf("expected value=NIl, got=%#v", u)
		}
		if ht.size() != uint64(i+1) {
			t.Fatalf("expected ht.size()=%d, got=%d", i+1, ht.size())
		}
	}

	for i := 0; i < 100; i++ {
		k := String(fmt.Sprintf("key:%d", i))
		mustDelete(t, ht, k)
		if ht.size() != uint64(100-(i+1)) {
			t.Fatalf("expected ht.size()=%d, got=%d", 100-(i+1), ht.size())
		}
		value, found, err := ht.get(k)
		if err != nil {
			t.Fatalf("expected err == nil, got=%#v", err)
		}
		if found {
			t.Fatalf("expected key=%#v to not be in hash table", k)
		}
		if value != nil {
			t.Fatalf("expected value == nil, got=%#v", err)
		}
		mustInsert(t, ht, k, NIL)
		if ht.size() != uint64(100-i) {
			t.Fatalf("expected ht.size()=%d, got=%d", 100-i, ht.size())
		}
		mustDelete(t, ht, k)
	}
}

func TestHashTableKeyTypes(t *testing.T) {
	pairs := []struct {
		k Hashable
		v Value
	}{
		{NIL, newObject(nil)},
		{TRUE, newObject(nil)},
		{FALSE, newObject(nil)},
		{Number(1), newObject(nil)},
		{Number(1.5), newObject(nil)},
		{String("a"), newObject(nil)},
		{String("b"), newObject(nil)},
	}
	ctx := NewContext()
	ht := newHashTable(ctx)
	for i, pair := range pairs {
		mustInsert(t, ht, pair.k, pair.v)
		v := mustGet(t, ht, pair.k)
		if v != pair.v {
			t.Fatalf("tests[%d] ht.get: expected=%p, got=%p", i, pair.v, v)
		}
		if got := ht.size(); got != uint64(i + 1) {
			t.Fatalf("tests[%d] ht.size(): expected=%d, got=%d", i, i + 1, got)
		}
	}
}

func mustInsert(t *testing.T, ht *hashTable, k Hashable, v Value) {
	if err := ht.insert(k, v); err != nil {
		t.Fatalf("unexpected insertion error=%#v", err)
	}
}

func mustGet(t *testing.T, ht *hashTable, k Hashable) Value {
	value, found, err := ht.get(k)
	if !found {
		t.Fatalf("expected key %#v to be in hash table", k)
	}
	if err != nil {
		t.Fatalf("unexpected get error=%#v", value)
	}
	return value
}

func mustDelete(t *testing.T, ht *hashTable, k Hashable) {
	found, err := ht.delete(k)
	if !found {
		if err != nil {
			t.Fatalf("unexpected delete error=%#v", err)
		}
		t.Fatalf("expected key %#v to be in hash table", k)
	}
}

// ----------
// Benchmarks
// ----------

func BenchmarkHashNil(b *testing.B) {
	// this involves no resizes, so is a good test to see how fast
	// the get and insert operations are.
	ctx := NewContext()
	ht := newHashTable(ctx)
	for n := 0; n < b.N; n++ {
		ht.insert(NIL, NIL)
		ht.get(NIL)
	}
}

func BenchmarkHashStrings(b *testing.B) {
	ctx := NewContext()
	ht := newHashTable(ctx)
	for n := 0; n < b.N; n++ {
		v := String(fmt.Sprintf("key:%d", n))
		ht.insert(v, nil)
		ht.get(v)
	}
	for n := 0; n < b.N; n++ {
		v := String(fmt.Sprintf("key:%d", n))
		ht.delete(v)
	}
}

func BenchmarkHashNumbers(b *testing.B) {
	ctx := NewContext()
	ht := newHashTable(ctx)
	for n := 0; n < b.N; n++ {
		v := Number(n)
		ht.insert(v, nil)
		ht.get(v)
	}
	for n := 0; n < b.N; n++ {
		ht.delete(Number(n))
	}
}
