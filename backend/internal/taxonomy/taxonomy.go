package taxonomy

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"organizer/backend/internal/domain"
)

//go:embed taxonomy.json
var defaultTaxonomyJSON []byte

// Rule represents a declarative classification rule.
type Rule struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"subCategory"`
	Patterns    []string `json:"patterns"`
}

// Config wraps the JSON array or object of rules.
type Config struct {
	Rules []Rule `json:"rules"`
}

type compiledRule struct {
	rule     Rule
	regexes  []*regexp.Regexp
}

// Engine implements declarative item classification against taxonomy rules.
type Engine struct {
	mu             sync.RWMutex
	rules          []Rule
	compiled       []compiledRule
	customFilePath string
}

// NewEngine initializes a taxonomy Engine with priority:
// 1. Local ./taxonomy.json
// 2. ~/.config/organizer/taxonomy.json
// 3. Embedded taxonomy.json
func NewEngine() *Engine {
	engine := &Engine{}

	// 1. Check local ./taxonomy.json
	if fileExists("./taxonomy.json") {
		if err := engine.loadFromFile("./taxonomy.json"); err == nil {
			return engine
		}
	}

	// 2. Check ~/.config/organizer/taxonomy.json
	if home, err := os.UserHomeDir(); err == nil {
		userConfig := filepath.Join(home, ".config", "organizer", "taxonomy.json")
		if fileExists(userConfig) {
			if err := engine.loadFromFile(userConfig); err == nil {
				return engine
			}
		}
	}

	// 3. Fallback to embedded default taxonomy
	if err := engine.loadFromBytes(defaultTaxonomyJSON, ""); err != nil {
		// As ultimate fallback, initialize empty
		engine.rules = make([]Rule, 0)
	}
	return engine
}

// NewEngineWithFile initializes an engine from a specific file path.
func NewEngineWithFile(path string) (*Engine, error) {
	engine := &Engine{}
	if err := engine.loadFromFile(path); err != nil {
		return nil, err
	}
	return engine, nil
}

// NewDefaultEngine creates a taxonomy Engine strictly initialized from embedded default rules.
func NewDefaultEngine() *Engine {
	engine := &Engine{}
	if err := engine.loadFromBytes(defaultTaxonomyJSON, ""); err != nil {
		engine.rules = make([]Rule, 0)
	}
	return engine
}

// SetCustomFilePath explicitly configures the persistence path for custom rules.
func (e *Engine) SetCustomFilePath(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.customFilePath = path
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (e *Engine) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return e.loadFromBytes(data, path)
}

func (e *Engine) loadFromBytes(data []byte, sourcePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var rules []Rule
	// Try parsing as Config object { "rules": [...] }
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil && len(cfg.Rules) > 0 {
		rules = cfg.Rules
	} else {
		// Try parsing as direct array [ ... ]
		if err := json.Unmarshal(data, &rules); err != nil {
			return fmt.Errorf("failed to parse taxonomy JSON: %w", err)
		}
	}

	e.rules = rules
	e.customFilePath = sourcePath
	e.compileRulesLocked()
	return nil
}

func (e *Engine) compileRulesLocked() {
	e.compiled = make([]compiledRule, 0, len(e.rules))
	for _, r := range e.rules {
		cr := compiledRule{
			rule:    r,
			regexes: make([]*regexp.Regexp, 0, len(r.Patterns)),
		}
		for _, pat := range r.Patterns {
			re := patternToRegex(pat)
			if re != nil {
				cr.regexes = append(cr.regexes, re)
			}
		}
		e.compiled = append(e.compiled, cr)
	}
}

// patternToRegex compiles a glob/string pattern into a boundary-aware regex.
func patternToRegex(pat string) *regexp.Regexp {
	pat = strings.TrimSpace(strings.ToLower(pat))
	if pat == "" {
		return nil
	}

	// Escape special regex characters except '*'
	var sb strings.Builder
	for i := 0; i < len(pat); i++ {
		c := pat[i]
		switch c {
		case '*':
			sb.WriteString(`[a-zA-Z0-9_-]*`)
		case '.', '+', '?', '(', ')', '[', ']', '{', '}', '^', '$', '|', '\\':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
		}
	}

	// Delimit by boundary: start/end of string or non-alphanumeric character
	// e.g. (: / - _ . space)
	patternExpr := fmt.Sprintf(`(?i)(?:^|[^a-zA-Z0-9])%s(?:$|[^a-zA-Z0-9])`, sb.String())
	re, err := regexp.Compile(patternExpr)
	if err != nil {
		return nil
	}
	return re
}

// Classify determines the Category, SubCategory, and IsClassified status for an Item.
// It inspects Name, ID, SourcePath, and Description.
func (e *Engine) Classify(item domain.Item) (string, string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	targets := []string{
		item.Name,
		item.ID,
		item.SourcePath,
		item.Description,
	}

	for _, cr := range e.compiled {
		for _, re := range cr.regexes {
			for _, target := range targets {
				if target != "" && re.MatchString(target) {
					return cr.rule.Category, cr.rule.SubCategory, true
				}
			}
		}
	}

	return "Unclassified", "Pending Triage", false
}

// ClassifyItem updates the Item's Category, SubCategory, and IsClassified fields in-place.
func (e *Engine) ClassifyItem(item *domain.Item) {
	cat, subCat, isClassified := e.Classify(*item)
	item.Category = cat
	item.SubCategory = subCat
	item.IsClassified = isClassified
}

// AddRule adds a new rule and persists it to the custom taxonomy file.
func (e *Engine) AddRule(category, subCategory string, patterns []string) error {
	category = strings.TrimSpace(category)
	subCategory = strings.TrimSpace(subCategory)
	if category == "" || subCategory == "" || len(patterns) == 0 {
		return fmt.Errorf("category, subCategory, and patterns are required")
	}

	cleanPatterns := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			cleanPatterns = append(cleanPatterns, p)
		}
	}
	if len(cleanPatterns) == 0 {
		return fmt.Errorf("no valid patterns provided")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if matching category/subCategory rule already exists
	updated := false
	for i, r := range e.rules {
		if strings.EqualFold(r.Category, category) && strings.EqualFold(r.SubCategory, subCategory) {
			existingMap := make(map[string]bool)
			for _, p := range r.Patterns {
				existingMap[strings.ToLower(p)] = true
			}
			for _, p := range cleanPatterns {
				if !existingMap[strings.ToLower(p)] {
					e.rules[i].Patterns = append(e.rules[i].Patterns, p)
				}
			}
			updated = true
			break
		}
	}

	if !updated {
		e.rules = append(e.rules, Rule{
			Category:    category,
			SubCategory: subCategory,
			Patterns:    cleanPatterns,
		})
	}

	e.compileRulesLocked()

	// Persist to custom taxonomy file
	return e.persistLocked()
}

func (e *Engine) persistLocked() error {
	savePath := e.customFilePath
	if savePath == "" {
		if fileExists("./taxonomy.json") {
			savePath = "./taxonomy.json"
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				savePath = "./taxonomy.json"
			} else {
				dir := filepath.Join(home, ".config", "organizer")
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create config dir: %w", err)
				}
				savePath = filepath.Join(dir, "taxonomy.json")
			}
		}
		e.customFilePath = savePath
	}

	cfg := Config{Rules: e.rules}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal taxonomy config: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write taxonomy file %s: %w", savePath, err)
	}

	return nil
}

// GetRules returns a copy of current rules.
func (e *Engine) GetRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Rule, len(e.rules))
	copy(result, e.rules)
	return result
}

// GetCategories returns unique category names.
func (e *Engine) GetCategories() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	seen := make(map[string]bool)
	var cats []string
	for _, r := range e.rules {
		if !seen[r.Category] {
			seen[r.Category] = true
			cats = append(cats, r.Category)
		}
	}
	sort.Strings(cats)
	return cats
}

// GetSubCategories returns unique subcategory names for a given category.
func (e *Engine) GetSubCategories(category string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	seen := make(map[string]bool)
	var subCats []string
	for _, r := range e.rules {
		if strings.EqualFold(r.Category, category) {
			if !seen[r.SubCategory] {
				seen[r.SubCategory] = true
				subCats = append(subCats, r.SubCategory)
			}
		}
	}
	sort.Strings(subCats)
	return subCats
}
