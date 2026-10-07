package uid

import (
	"regexp"
	"strings"
	"testing"
)

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestNew(t *testing.T) {
	id := New()
	if !hex32.MatchString(id) {
		t.Fatalf("New() = %q, want 32 lowercase hex chars", id)
	}
}

func TestNewUnique(t *testing.T) {
	seen := make(map[string]struct{}, 10000)
	for i := 0; i < 10000; i++ {
		id := New()
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestWithPrefix(t *testing.T) {
	id := WithPrefix("ord")
	if !strings.HasPrefix(id, "ord_") {
		t.Fatalf("WithPrefix(\"ord\") = %q, want prefix %q", id, "ord_")
	}
	if body := strings.TrimPrefix(id, "ord_"); !hex32.MatchString(body) {
		t.Fatalf("unexpected id body: %q", body)
	}
}

func TestWithPrefixEmpty(t *testing.T) {
	if id := WithPrefix(""); !hex32.MatchString(id) {
		t.Fatalf("WithPrefix(\"\") = %q, want plain 32 hex chars", id)
	}
}

func TestNewSortable(t *testing.T) {
	prev := NewSortable()
	if !hex32.MatchString(prev) {
		t.Fatalf("NewSortable() = %q, want 32 lowercase hex chars", prev)
	}
	// UUID v7 carries a version nibble of 7 at index 12.
	if prev[12] != '7' {
		t.Fatalf("NewSortable() = %q, want version nibble 7", prev)
	}
	for i := 0; i < 1000; i++ {
		id := NewSortable()
		if id <= prev {
			t.Fatalf("not monotonic: %s <= %s", id, prev)
		}
		prev = id
	}
}

func TestWithPrefixSortable(t *testing.T) {
	id := WithPrefixSortable("ord")
	if !strings.HasPrefix(id, "ord_") {
		t.Fatalf("WithPrefixSortable(\"ord\") = %q, want prefix %q", id, "ord_")
	}
	if body := strings.TrimPrefix(id, "ord_"); !hex32.MatchString(body) {
		t.Fatalf("unexpected id body: %q", body)
	}
	if id := WithPrefixSortable(""); !hex32.MatchString(id) {
		t.Fatalf("WithPrefixSortable(\"\") = %q, want plain 32 hex chars", id)
	}
}
