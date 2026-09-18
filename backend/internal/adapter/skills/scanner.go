package skills

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
)

// Target defines a directory to scan and its corresponding AI provider.
type Target struct {
	Path     string
	Provider domain.Provider
}

// DefaultTargets contains the canonical directories and providers specified for skills.
var DefaultTargets = []Target{
	{Path: "~/.agents/skills", Provider: domain.ProviderGlobal},
	{Path: "~/.cursor/skills", Provider: domain.ProviderCursor},
	{Path: "~/.config/opencode/skills", Provider: domain.ProviderOpenCode},
	{Path: "~/.grok/skills", Provider: domain.ProviderGrok},
	{Path: "~/.grok/bundled/skills", Provider: domain.ProviderGrok},
	{Path: "~/.hermes/skills", Provider: domain.ProviderHermes},
	{Path: "~/.pi/agent/skills", Provider: domain.ProviderPi},
}

const (
	DefaultHermesManifestPath = "~/.hermes/skills/.bundled_manifest"
	DefaultSkillLockPath      = "~/.agents/.skill-lock.json"
)

type skillLockFile struct {
	Skills map[string]skillLockEntry `json:"skills"`
}

type skillLockEntry struct {
	Source    string `json:"source"`
	SourceURL string `json:"sourceUrl"`
}

func loadHermesBundledManifest(path string) map[string]bool {
	manifest := make(map[string]bool)
	expanded := adapter.ExpandHome(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return manifest
	}
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		if key != "" {
			manifest[key] = true
		}
	}
	return manifest
}

func loadSkillLock(path string) map[string]skillLockEntry {
	lockMap := make(map[string]skillLockEntry)
	expanded := adapter.ExpandHome(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return lockMap
	}
	var file skillLockFile
	if err := json.Unmarshal(data, &file); err != nil {
		return lockMap
	}
	if file.Skills != nil {
		return file.Skills
	}
	return lockMap
}

// Scanner scans multiple directory targets for SKILL.md files.
type Scanner struct {
	targets            []Target
	hermesManifestPath string
	skillLockPath      string
}

// NewScanner creates a new skills scanner with given or default targets.
func NewScanner(targets ...Target) *Scanner {
	if len(targets) == 0 {
		targets = DefaultTargets
	}
	return &Scanner{
		targets:            targets,
		hermesManifestPath: DefaultHermesManifestPath,
		skillLockPath:      DefaultSkillLockPath,
	}
}

// WithManifestPaths overrides default manifest paths (useful for tests).
func (s *Scanner) WithManifestPaths(hermesManifest, skillLock string) *Scanner {
	if hermesManifest != "" {
		s.hermesManifestPath = hermesManifest
	}
	if skillLock != "" {
		s.skillLockPath = skillLock
	}
	return s
}

// Name returns the identifier of this adapter.
func (s *Scanner) Name() string {
	return "skills-scanner"
}

// Scan scans all configured skill target directories and extracts domain Items.
func (s *Scanner) Scan(ctx context.Context) ([]domain.Item, error) {
	var allItems []domain.Item
	seenIDs := make(map[string]bool)

	hermesManifest := loadHermesBundledManifest(s.hermesManifestPath)
	skillLock := loadSkillLock(s.skillLockPath)

	for _, target := range s.targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		expandedPath := adapter.ExpandHome(target.Path)
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			// Skip directories that do not exist gracefully
			continue
		}

		visited := make(map[string]bool)
		items, err := s.scanDir(expandedPath, target.Provider, visited, hermesManifest, skillLock)
		if err != nil {
			continue
		}

		for _, it := range items {
			if !seenIDs[it.ID] {
				seenIDs[it.ID] = true
				allItems = append(allItems, it)
			}
		}
	}

	return allItems, nil
}

func (s *Scanner) scanDir(dir string, provider domain.Provider, visited map[string]bool, hermesManifest map[string]bool, skillLock map[string]skillLockEntry) ([]domain.Item, error) {
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, nil
	}
	if visited[realDir] {
		return nil, nil
	}
	visited[realDir] = true

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	var items []domain.Item
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden directories/files (e.g. .git, .archive)
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if strings.EqualFold(name, "SKILL.md") {
			item, err := parseSkillFile(fullPath, provider, hermesManifest, skillLock)
			if err == nil {
				items = append(items, item)
			}
			continue
		}

		stat, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if stat.IsDir() {
			subItems, err := s.scanDir(fullPath, provider, visited, hermesManifest, skillLock)
			if err == nil {
				items = append(items, subItems...)
			}
		}
	}

	return items, nil
}

type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func parseSkillFile(skillPath string, provider domain.Provider, hermesManifest map[string]bool, skillLock map[string]skillLockEntry) (domain.Item, error) {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return domain.Item{}, err
	}

	rawText := string(content)
	fm, bodyPreview, rawConfig := extractFrontmatterAndBody(rawText)

	folderName := filepath.Base(filepath.Dir(skillPath))
	name := strings.TrimSpace(fm.Name)
	if name == "" {
		// Fallback to parent directory name
		name = folderName
	}

	desc := strings.TrimSpace(fm.Description)
	if desc == "" {
		desc = bodyPreview
	}

	id := fmt.Sprintf("%s:%s:%s", provider, domain.ItemTypeSkill, name)

	origin, sourceURL, sourcePkg, registryURL := classifySkillOrigin(skillPath, provider, name, folderName, hermesManifest, skillLock)

	return domain.Item{
		ID:          id,
		Name:        name,
		Type:        domain.ItemTypeSkill,
		Provider:    provider,
		Origin:      origin,
		SourcePath:  skillPath,
		Description: desc,
		SourceURL:   sourceURL,
		SourcePkg:   sourcePkg,
		RegistryURL: registryURL,
		Invocation:  adapter.FormatSkillInvocation(name),
		RawConfig:   rawConfig,
	}, nil
}

func classifySkillOrigin(
	skillPath string,
	provider domain.Provider,
	name string,
	folderName string,
	hermesManifest map[string]bool,
	skillLock map[string]skillLockEntry,
) (string, string, string, string) {
	normalizedPath := filepath.ToSlash(skillPath)

	// 1. Grok:
	// If under bundled/skills, it is "builtin".
	// If under ~/.grok/skills, it is "custom".
	if provider == domain.ProviderGrok {
		if strings.Contains(normalizedPath, "bundled/skills") {
			return domain.OriginBuiltin, "", "", ""
		}
		return domain.OriginCustom, "", "", ""
	}

	// 2. Hermes:
	// If a Hermes skill folder name is present as a key in this file (e.g. airtable:30f4...), its origin is "builtin".
	// If not in that file, its origin is "custom" (or installed if in .skill-lock.json).
	if provider == domain.ProviderHermes {
		if hermesManifest[folderName] || hermesManifest[name] {
			return domain.OriginBuiltin, "", "", ""
		}
		if entry, ok := skillLock[name]; ok {
			sURL := entry.SourceURL
			if sURL == "" {
				sURL = entry.Source
			}
			regURL := ""
			if entry.Source != "" {
				regURL = fmt.Sprintf("https://skills.sh/%s/%s", entry.Source, name)
			}
			return domain.OriginInstalled, sURL, entry.Source, regURL
		}
		if entry, ok := skillLock[folderName]; ok {
			sURL := entry.SourceURL
			if sURL == "" {
				sURL = entry.Source
			}
			regURL := ""
			if entry.Source != "" {
				regURL = fmt.Sprintf("https://skills.sh/%s/%s", entry.Source, folderName)
			}
			return domain.OriginInstalled, sURL, entry.Source, regURL
		}
		return domain.OriginCustom, "", "", ""
	}

	// 3. Global / Skills CLI & Other providers:
	// If skill name matches a key in .skill-lock.json, its origin is "installed", populate SourceURL with sourceUrl or source.
	// If not in that file, its origin is "custom".
	if entry, ok := skillLock[name]; ok {
		sURL := entry.SourceURL
		if sURL == "" {
			sURL = entry.Source
		}
		regURL := ""
		if entry.Source != "" {
			regURL = fmt.Sprintf("https://skills.sh/%s/%s", entry.Source, name)
		}
		return domain.OriginInstalled, sURL, entry.Source, regURL
	}
	if entry, ok := skillLock[folderName]; ok {
		sURL := entry.SourceURL
		if sURL == "" {
			sURL = entry.Source
		}
		regURL := ""
		if entry.Source != "" {
			regURL = fmt.Sprintf("https://skills.sh/%s/%s", entry.Source, folderName)
		}
		return domain.OriginInstalled, sURL, entry.Source, regURL
	}

	return domain.OriginCustom, "", "", ""
}

func extractFrontmatterAndBody(content string) (frontmatter, string, string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var fm frontmatter
	var rawConfig string
	bodyStartLine := 0

	// Check for leading ---
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		closingIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closingIdx = i
				break
			}
		}

		if closingIdx != -1 {
			yamlLines := lines[1:closingIdx]
			yamlContent := strings.Join(yamlLines, "\n")
			rawConfig = yamlContent
			_ = yaml.Unmarshal([]byte(yamlContent), &fm)
			bodyStartLine = closingIdx + 1
		}
	}

	// Extract body preview from lines after frontmatter
	var bodyLines []string
	for i := bodyStartLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Strip markdown header prefix for cleaner preview
		cleanLine := strings.TrimLeft(line, "# ")
		if cleanLine != "" {
			bodyLines = append(bodyLines, cleanLine)
		}
		if len(strings.Join(bodyLines, " ")) >= 300 {
			break
		}
	}

	bodyPreview := strings.Join(bodyLines, " ")
	if len(bodyPreview) > 300 {
		bodyPreview = strings.TrimSpace(bodyPreview[:300]) + "..."
	}

	return fm, bodyPreview, rawConfig
}
