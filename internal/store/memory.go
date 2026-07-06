package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/9KMan/JOB-20260706144427-000133/internal/model"
	"github.com/google/uuid"
)

// ErrEmptyName is returned when an item name is empty or whitespace.
var ErrEmptyName = errors.New("name cannot be empty")

// ErrNotFound is returned when an item is not found.
var ErrNotFound = errors.New("item not found")

// ItemStore defines the interface for item storage operations.
type ItemStore interface {
	List() []model.Item
	Create(name string) (model.Item, error)
	Get(id string) (model.Item, error)
}

// MemoryStore implements ItemStore using in-memory storage.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]model.Item
}

// NewMemoryStore creates a new MemoryStore with pre-seeded items.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		items: make(map[string]model.Item),
	}
	// Pre-seed with 2 items
	store.items["1"] = model.Item{
		ID:        "1",
		Name:      "Welcome to Grux",
		CreatedAt: time.Now().UTC(),
	}
	store.items["2"] = model.Item{
		ID:        "2",
		Name:      "Hello, world!",
		CreatedAt: time.Now().UTC(),
	}
	return store
}

// List returns all items in the store.
func (s *MemoryStore) List() []model.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]model.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

// Create creates a new item with the given name. Empty or whitespace-only
// names return ErrEmptyName; trimmed names are stored verbatim.
func (s *MemoryStore) Create(name string) (model.Item, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return model.Item{}, ErrEmptyName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item := model.Item{
		ID:        uuid.New().String(),
		Name:      trimmed,
		CreatedAt: time.Now().UTC(),
	}
	s.items[item.ID] = item
	return item, nil
}

// Get retrieves an item by its ID.
func (s *MemoryStore) Get(id string) (model.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return model.Item{}, ErrNotFound
	}
	return item, nil
}
