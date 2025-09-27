package todo

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/memstore"
)

const (
	schemaKey           = "meta:schema_version"
	SchemaVersion       = "1"
	itemKeyPrefix       = "item:"
	namespacePrefixBase = "todo:project:"
)

var (
	ErrNotFound      = errors.New("todo: item not found")
	ErrEmptyTitle    = errors.New("todo: title is required")
	ErrSchemaUnknown = errors.New("todo: unsupported schema version")
)

// Item represents a stored TODO record and doubles as a Bubble list.Item.
type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    string    `json:"priority,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	AgentID     string    `json:"agent_id,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Source      string    `json:"source,omitempty"`
}

// FilterValue allows Item to satisfy list.Item in the TUI.
func (i Item) FilterValue() string { return i.Title }

// Filter narrows the items returned by List.
type Filter struct {
	Status   string
	Priority string
	Tags     []string
}

// CreateParams defines the fields accepted when creating a new item.
type CreateParams struct {
	Title       string
	Description string
	Priority    string
	Tags        []string
	AgentID     string
	Source      string
	Status      string
}

// UpdateParams defines optional updates for an existing item.
type UpdateParams struct {
	Title       *string
	Description *string
	Priority    *string
	Tags        *[]string
	AgentID     *string
	Status      *string
}

// Service provides access to todo storage in the shared memstore.
type Service struct {
	store     memstore.SharedStore
	namespace string
}

// NewService constructs a Service for the provided workspace root (or current dir).
func NewService(store memstore.SharedStore, projectPath string) (*Service, error) {
	if store == nil {
		store = memstore.Get()
	}

	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("todo: resolve working directory: %w", err)
		}
		projectPath = cwd
	}

	abs, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("todo: resolve project path: %w", err)
	}

	ns := namespaceForPath(abs)
	svc := &Service{store: store, namespace: ns}
	if err := svc.ensureSchemaVersion(); err != nil {
		return nil, err
	}
	return svc, nil
}

// namespaceForPath hashes a filesystem path to derive a stable namespace.
func namespaceForPath(path string) string {
	sum := sha1.Sum([]byte(path))
	return namespacePrefixBase + hex.EncodeToString(sum[:8])
}

// Namespace exposes the underlying memstore namespace (primarily for tests).
func (s *Service) Namespace() string { return s.namespace }

// ensureSchemaVersion records and validates the schema version for the namespace.
func (s *Service) ensureSchemaVersion() error {
	payload := []byte(SchemaVersion)
	b, ok, err := s.store.Get(s.namespace, schemaKey)
	if err != nil {
		return fmt.Errorf("todo: load schema version: %w", err)
	}
	if !ok {
		if err := s.store.Set(s.namespace, schemaKey, payload, 0); err != nil {
			return fmt.Errorf("todo: write schema version: %w", err)
		}
		return nil
	}
	stored := strings.TrimSpace(string(b))
	if stored != SchemaVersion {
		return fmt.Errorf("%w: have %s want %s", ErrSchemaUnknown, stored, SchemaVersion)
	}
	return nil
}

// Create inserts a new item using the provided params.
func (s *Service) Create(params CreateParams) (Item, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return Item{}, ErrEmptyTitle
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("%d", now.UnixNano())
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "pending"
	}

	item := Item{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(params.Description),
		Priority:    strings.TrimSpace(params.Priority),
		Tags:        normalizeTags(params.Tags),
		AgentID:     strings.TrimSpace(params.AgentID),
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
		Source:      strings.TrimSpace(params.Source),
	}

	if err := s.save(item); err != nil {
		return Item{}, err
	}
	return item, nil
}

// List returns every item that matches the filter criteria.
func (s *Service) List(filter Filter) ([]Item, error) {
	keys, err := s.store.Keys(s.namespace)
	if err != nil {
		return nil, fmt.Errorf("todo: list keys: %w", err)
	}

	var items []Item
	for _, key := range keys {
		if !strings.HasPrefix(key, itemKeyPrefix) {
			continue
		}
		it, ok, err := s.load(key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if !filter.matches(it) {
			continue
		}
		items = append(items, it)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

// Get fetches a single item by id.
func (s *Service) Get(id string) (Item, bool, error) {
	if strings.TrimSpace(id) == "" {
		return Item{}, false, nil
	}
	key := keyForID(id)
	it, ok, err := s.load(key)
	if err != nil {
		return Item{}, false, err
	}
	return it, ok, nil
}

// Update applies the supplied fields to an existing item.
func (s *Service) Update(id string, updates UpdateParams) (Item, error) {
	key := keyForID(id)
	it, ok, err := s.load(key)
	if err != nil {
		return Item{}, err
	}
	if !ok {
		return Item{}, ErrNotFound
	}

	if updates.Title != nil {
		trimmed := strings.TrimSpace(*updates.Title)
		if trimmed == "" {
			return Item{}, ErrEmptyTitle
		}
		it.Title = trimmed
	}
	if updates.Description != nil {
		it.Description = strings.TrimSpace(*updates.Description)
	}
	if updates.Priority != nil {
		it.Priority = strings.TrimSpace(*updates.Priority)
	}
	if updates.Tags != nil {
		it.Tags = normalizeTags(*updates.Tags)
	}
	if updates.AgentID != nil {
		it.AgentID = strings.TrimSpace(*updates.AgentID)
	}
	if updates.Status != nil {
		trimmed := strings.TrimSpace(*updates.Status)
		if trimmed != "" {
			it.Status = trimmed
		}
	}
	it.UpdatedAt = time.Now().UTC()

	if err := s.save(it); err != nil {
		return Item{}, err
	}
	return it, nil
}

// Delete removes an item by id.
func (s *Service) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrNotFound
	}
	if err := s.store.Delete(s.namespace, keyForID(id)); err != nil {
		return fmt.Errorf("todo: delete item: %w", err)
	}
	return nil
}

func (s *Service) save(item Item) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("todo: marshal item: %w", err)
	}
	if err := s.store.Set(s.namespace, keyForID(item.ID), data, 0); err != nil {
		return fmt.Errorf("todo: store item: %w", err)
	}
	return nil
}

func (s *Service) load(key string) (Item, bool, error) {
	b, ok, err := s.store.Get(s.namespace, key)
	if err != nil {
		return Item{}, false, fmt.Errorf("todo: load item: %w", err)
	}
	if !ok {
		return Item{}, false, nil
	}
	var it Item
	if err := json.Unmarshal(b, &it); err != nil {
		return Item{}, false, fmt.Errorf("todo: decode item: %w", err)
	}
	return it, true, nil
}

func keyForID(id string) string {
	return itemKeyPrefix + id
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	var out []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func (f Filter) matches(it Item) bool {
	if f.Status != "" && it.Status != f.Status {
		return false
	}
	if f.Priority != "" && it.Priority != f.Priority {
		return false
	}
	if len(f.Tags) > 0 && !hasAllTags(it.Tags, f.Tags) {
		return false
	}
	return true
}

func hasAllTags(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	if len(have) == 0 {
		return false
	}
	lookup := make(map[string]struct{}, len(have))
	for _, tag := range have {
		lookup[strings.ToLower(tag)] = struct{}{}
	}
	for _, tag := range want {
		if _, ok := lookup[strings.ToLower(tag)]; !ok {
			return false
		}
	}
	return true
}
