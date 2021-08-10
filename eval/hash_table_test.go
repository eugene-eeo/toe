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
		err := ht.insert(k, v)
		if err != nil {
			t.Fatalf("unexpected insertion error=%#+v", err)
		}
		u, found := ht.get(String(fmt.Sprintf("key:%d", i)))
		if !found {
			t.Fatalf("expected key %#+v to be in hash table", k)
			if isError(u) {
				t.Errorf("error retrieving key %#+v:\n", k)
				t.Fatalf("  %#+v\n", u)
			}
		}
		if u != v {
			t.Fatalf("expected=%#+v, got=%#+v", v, u)
		}
		if ht.size() != uint64(i+1) {
			t.Fatalf("expected ht.size()=%d, got=%d", i+1, ht.size())
		}
	}

	for i := 0; i < 100; i++ {
		k := String(fmt.Sprintf("key:%d", i))
		found, err := ht.delete(k)
		if err != nil {
			t.Fatalf("unexpected deletion error=%#v", err)
		}
		if !found {
			t.Fatalf("deletion of key=%#v failed", k)
		}
		if ht.size() != uint64(100-(i+1)) {
			t.Fatalf("expected ht.size()=%d, got=%d", 100-(i+1), ht.size())
		}
		u, found := ht.get(k)
		if found {
			t.Fatalf("expected key=%#v to not be in hash table", k)
		}
		if u != nil {
			t.Fatalf("expected value == nil, got=%#v", u)
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
