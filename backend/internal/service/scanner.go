package service

import (
	"context"
	"strings"
	"sync"

	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
	"organizer/backend/internal/taxonomy"
)

// AggregatorService coordinates all scanner adapters and maintains scanned items.
type AggregatorService struct {
	adapters []adapter.Adapter
	taxonomy *taxonomy.Engine
	mu       sync.RWMutex
	items    []domain.Item
}

// NewAggregatorService creates a new AggregatorService with registered adapters.
func NewAggregatorService(adapters ...adapter.Adapter) *AggregatorService {
	return &AggregatorService{
		adapters: adapters,
		taxonomy: taxonomy.NewEngine(),
		items:    make([]domain.Item, 0),
	}
}

// RegisterAdapter adds an adapter to the service.
func (s *AggregatorService) RegisterAdapter(a adapter.Adapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adapters = append(s.adapters, a)
}

// SetTaxonomy sets a custom taxonomy engine and reclassifies all items.
func (s *AggregatorService) SetTaxonomy(t *taxonomy.Engine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taxonomy = t
	for i := range s.items {
		s.taxonomy.ClassifyItem(&s.items[i])
	}
	s.items = taxonomy.InferFamilies(s.items)
}

// GetTaxonomy returns the current taxonomy engine.
func (s *AggregatorService) GetTaxonomy() *taxonomy.Engine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.taxonomy
}

// ReclassifyAll re-runs taxonomy classification on all cached items.
func (s *AggregatorService) ReclassifyAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		s.taxonomy.ClassifyItem(&s.items[i])
	}
	s.items = taxonomy.InferFamilies(s.items)
}

// ScanAll runs all registered adapters, aggregates results, deduplicates by ID, classifies items, and caches them.
func (s *AggregatorService) ScanAll(ctx context.Context) ([]domain.Item, error) {
	s.mu.RLock()
	adapterList := make([]adapter.Adapter, len(s.adapters))
	copy(adapterList, s.adapters)
	taxEngine := s.taxonomy
	s.mu.RUnlock()

	var wg sync.WaitGroup
	var collectMu sync.Mutex
	var collected []domain.Item
	seen := make(map[string]bool)

	for _, adp := range adapterList {
		wg.Add(1)
		go func(a adapter.Adapter) {
			defer wg.Done()
			items, err := a.Scan(ctx)
			if err != nil {
				// Adapter errors are non-fatal to allow partial results
				return
			}

			collectMu.Lock()
			defer collectMu.Unlock()
			for _, it := range items {
				if !seen[it.ID] {
					seen[it.ID] = true
					// Run taxonomy classification
					taxEngine.ClassifyItem(&it)
					collected = append(collected, it)
				}
			}
		}(adp)
	}

	wg.Wait()

	collected = taxonomy.InferFamilies(collected)

	s.mu.Lock()
	s.items = make([]domain.Item, len(collected))
	copy(s.items, collected)
	s.mu.Unlock()

	return collected, nil
}

// GetItems filters and searches currently cached items.
func (s *AggregatorService) GetItems(
	itemType domain.ItemType,
	provider domain.Provider,
	origin string,
	category string,
	subCategory string,
	family string,
	isClassified *bool,
	search string,
) []domain.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	search = strings.TrimSpace(strings.ToLower(search))
	targetType := strings.TrimSpace(string(itemType))
	targetProvider := strings.TrimSpace(string(provider))
	targetOrigin := strings.TrimSpace(strings.ToLower(origin))
	targetCategory := strings.TrimSpace(category)
	targetSubCategory := strings.TrimSpace(subCategory)
	targetFamily := strings.TrimSpace(strings.ToLower(family))

	results := make([]domain.Item, 0)

	for _, it := range s.items {
		// Filter by Type
		if targetType != "" && !strings.EqualFold(string(it.Type), targetType) {
			continue
		}

		// Filter by Provider
		if targetProvider != "" && !strings.EqualFold(string(it.Provider), targetProvider) {
			continue
		}

		// Filter by Origin
		if targetOrigin != "" && !strings.EqualFold(it.Origin, targetOrigin) {
			continue
		}

		// Filter by Category
		if targetCategory != "" && !strings.EqualFold(it.Category, targetCategory) {
			continue
		}

		// Filter by SubCategory
		if targetSubCategory != "" && !strings.EqualFold(it.SubCategory, targetSubCategory) {
			continue
		}

		// Filter by Family
		if targetFamily != "" && !strings.EqualFold(it.Family, targetFamily) {
			continue
		}

		// Filter by IsClassified
		if isClassified != nil && it.IsClassified != *isClassified {
			continue
		}

		// Search term match across multiple fields
		if search != "" && !matchesSearch(it, search) {
			continue
		}

		results = append(results, it)
	}

	return results
}

// GetItemByID returns an item by its unique ID.
func (s *AggregatorService) GetItemByID(id string) (domain.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, it := range s.items {
		if it.ID == id {
			return it, true
		}
	}
	return domain.Item{}, false
}

// GetStats calculates summary statistics across cached items.
func (s *AggregatorService) GetStats() domain.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := domain.Stats{
		Total:             len(s.items),
		ByType:            make(map[domain.ItemType]int),
		ByProvider:        make(map[domain.Provider]int),
		ByOrigin:          make(map[string]int),
		ByCategory:        make(map[string]int),
		BySubCategory:     make(map[string]int),
		ByFamily:          make(map[string]int),
		ClassifiedCount:   0,
		UnclassifiedCount: 0,
	}

	for _, it := range s.items {
		stats.ByType[it.Type]++
		stats.ByProvider[it.Provider]++
		originKey := it.Origin
		if originKey == "" {
			originKey = domain.OriginCustom
		}
		stats.ByOrigin[originKey]++

		if it.IsClassified {
			stats.ClassifiedCount++
		} else {
			stats.UnclassifiedCount++
		}

		if it.Category != "" {
			stats.ByCategory[it.Category]++
		}
		if it.SubCategory != "" {
			stats.BySubCategory[it.SubCategory]++
		}
		if it.Family != "" {
			stats.ByFamily[it.Family]++
		}
	}

	return stats
}

// SetItems replaces the in-memory items (primarily for testing) and classifies them.
func (s *AggregatorService) SetItems(items []domain.Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make([]domain.Item, len(items))
	for i, it := range items {
		if s.taxonomy != nil && it.Category == "" {
			s.taxonomy.ClassifyItem(&it)
		}
		s.items[i] = it
	}
	s.items = taxonomy.InferFamilies(s.items)
}

func matchesSearch(it domain.Item, query string) bool {
	if strings.Contains(strings.ToLower(it.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Family), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Category), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.SubCategory), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Description), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Command), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Invocation), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.SourcePath), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.Origin), query) {
		return true
	}
	if strings.Contains(strings.ToLower(it.SourceURL), query) {
		return true
	}
	for _, arg := range it.Args {
		if strings.Contains(strings.ToLower(arg), query) {
			return true
		}
	}
	for _, key := range it.EnvKeys {
		if strings.Contains(strings.ToLower(key), query) {
			return true
		}
	}
	return false
}
