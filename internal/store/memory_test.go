package store

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestNewMemoryStoreSeeded(t *testing.T) {
	s := NewMemoryStore()
	items := s.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 seeded items, got %d", len(items))
	}
	names := make(map[string]bool)
	for _, it := range items {
		if it.ID == "" {
			t.Errorf("seeded item %q missing ID", it.Name)
		}
		if it.CreatedAt.IsZero() {
			t.Errorf("seeded item %q missing CreatedAt", it.Name)
		}
		names[it.Name] = true
	}
	if !names["Welcome to Grux"] || !names["Hello, world!"] {
		t.Errorf("expected seed names to be present, got %v", names)
	}
}

func TestMemoryStoreCreateSuccess(t *testing.T) {
	s := NewMemoryStore()
	item, err := s.Create("alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "alpha" {
		t.Errorf("expected name=alpha, got %q", item.Name)
	}
	if _, err := uuid.Parse(item.ID); err != nil {
		t.Errorf("expected UUID v4 ID, got %q (err=%v)", item.ID, err)
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestMemoryStoreCreateRejectsEmptyName(t *testing.T) {
	s := NewMemoryStore()
	cases := map[string]string{
		"empty":      "",
		"spaces":     "   ",
		"tabs":       "\t",
		"newlines":   "\n",
		"mixed":      " \t\n ",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := s.Create(value)
			if !errors.Is(err, ErrEmptyName) {
				t.Errorf("expected ErrEmptyName for %q, got %v", value, err)
			}
		})
	}
}

func TestMemoryStoreGetSuccess(t *testing.T) {
	s := NewMemoryStore()
	created, err := s.Create("beta")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Errorf("mismatch: got=%+v want=%+v", got, created)
	}
}

func TestMemoryStoreGetNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Get("missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStoreListReflectsCreates(t *testing.T) {
	s := NewMemoryStore()
	start := len(s.List())
	if _, err := s.Create("a"); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if _, err := s.Create("b"); err != nil {
		t.Fatalf("create b: %v", err)
	}
	if got := len(s.List()); got != start+2 {
		t.Errorf("expected %d items after 2 creates, got %d", start+2, got)
	}
}

// TestMemoryStoreConcurrentCreates verifies the sync.RWMutex protects the
// underlying map under concurrent writes. With the seed count of 2 plus
// n successful creates, the final list size must equal 2 + n.
func TestMemoryStoreConcurrentCreates(t *testing.T) {
	s := NewMemoryStore()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			name := "concurrent-" + strings.Repeat("x", i%8)
			if _, err := s.Create(name); err != nil {
				t.Errorf("concurrent create %d failed: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	if got := len(s.List()); got != 2+n {
		t.Errorf("expected %d items after concurrent creates, got %d", 2+n, got)
	}
}
