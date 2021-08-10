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
		// err := ht.insert(Number(i), v)
		err := ht.insert(k, v)
		if err != nil {
			t.Fatalf("unexpected insertion error=%#+v", err)
		}
		// u := ht.get(Number(i))
		u := ht.get(String(fmt.Sprintf("key:%d", i)))
		if u == nil {
			t.Fatalf("expected key %#+v to be in hash table", Number(i))
		}
		if isError(u) {
			t.Errorf("error retrieving key %#+v:\n", Number(i))
			t.Fatalf("  %#+v\n", u)
		}
		if u != v {
			t.Fatalf("expected=%#+v, got=%#+v", v, u)
		}
		if ht.sz != uint64(i+1) {
			t.Fatalf("expected ht.sz=%d, got=%d", i+1, ht.sz)
		}
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
}

func BenchmarkHashNumbers(b *testing.B) {
	ctx := NewContext()
	ht := newHashTable(ctx)
	for n := 0; n < b.N; n++ {
		v := Number(n)
		ht.insert(v, nil)
		ht.get(v)
	}
}
