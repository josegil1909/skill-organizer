package taxonomy

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
	"strings"

	"organizer/backend/internal/domain"
)

//go:embed packs.json
var defaultPacksJSON []byte

// CommonFamilies defines recognized family prefixes / namespaces.
// Sorted with multi-hyphen namespaces matching first.
var CommonFamilies = []string{
	"image-to-code",
	"playwright",
	"redteam",
	"golang",
	"resume",
	"baoyu",
	"taste",
	"skill",
	"osint",
	"recon",
	"game",
	"hunt",
	"a11y",
	"vue",
	"sdd",
	"cve",
	"bb",
}

type packAlias struct {
	Family   string   `json:"family"`
	Parent   string   `json:"parent"`
	Patterns []string `json:"patterns"`
}

type packConfig struct {
	Aliases []packAlias `json:"aliases"`
}

var loadedPackAliases []packAlias

func init() {
	var cfg packConfig
	if err := json.Unmarshal(defaultPacksJSON, &cfg); err == nil {
		loadedPackAliases = cfg.Aliases
	}
}

// InferFamilies processes a slice of items, detects family groupings,
// maps canonical parents (IsParent = true, ChildCount = N),
// and assigns Family slugs.
//
// Assignment priority (later wins):
//  1. Name prefix (existing CommonFamilies / hyphen groups)
//  2. Nested directory under /skills/ (Hermes-style packs)
//  3. Shared sourcePkg (skills.sh owner/repo)
//  4. Pack aliases when the declared parent skill is present (bug-bounty → hunt-*)
func InferFamilies(items []domain.Item) []domain.Item {
	if len(items) == 0 {
		return items
	}

	families := assignPrefixFamilies(items)
	overlayDirectoryPacks(items, families)
	overlaySourcePkgPacks(items, families)
	overlayAliasPacks(items, families)
	return stampParents(items, families)
}

func assignPrefixFamilies(items []domain.Item) []string {
	isCommonFam := make(map[string]bool)
	for _, f := range CommonFamilies {
		isCommonFam[f] = true
	}

	normNames := make([]string, len(items))
	allNames := make(map[string]bool)
	for i, it := range items {
		n := strings.TrimSpace(strings.ToLower(it.Name))
		normNames[i] = n
		if n != "" {
			allNames[n] = true
		}
	}

	isStandaloneParent := make(map[string]bool)
	for parent := range allNames {
		if len(parent) < 2 {
			continue
		}
		prefixHyphen := parent + "-"
		prefixUnderscore := parent + "_"
		for _, name := range normNames {
			if strings.HasPrefix(name, prefixHyphen) || strings.HasPrefix(name, prefixUnderscore) {
				isStandaloneParent[parent] = true
				break
			}
		}
	}

	candidateFamilies := make([]string, len(items))
	familyItemCounts := make(map[string]int)

	for i, name := range normNames {
		if name == "" {
			continue
		}

		var cand string

		for _, fam := range CommonFamilies {
			if name == fam || strings.HasPrefix(name, fam+"-") || strings.HasPrefix(name, fam+"_") {
				cand = fam
				break
			}
		}

		if cand == "" {
			for parent := range isStandaloneParent {
				if name == parent || strings.HasPrefix(name, parent+"-") || strings.HasPrefix(name, parent+"_") {
					cand = parent
					break
				}
			}
		}

		if cand == "" {
			parts := strings.Split(name, "-")
			if len(parts) > 1 && len(parts[0]) >= 2 {
				cand = parts[0]
			}
		}

		candidateFamilies[i] = cand
		if cand != "" {
			familyItemCounts[cand]++
		}
	}

	validFamilies := make(map[string]bool)
	for fam, count := range familyItemCounts {
		if isCommonFam[fam] || isStandaloneParent[fam] || count >= 2 {
			validFamilies[fam] = true
		}
	}

	for i, fam := range candidateFamilies {
		if !validFamilies[fam] {
			candidateFamilies[i] = ""
		}
	}
	return candidateFamilies
}

func overlayDirectoryPacks(items []domain.Item, families []string) {
	type pathInfo struct {
		segments []string
	}
	infos := make([]pathInfo, len(items))
	packCount := make(map[string]int)
	packHasNested := make(map[string]bool)

	for i, it := range items {
		segs := skillsRelSegments(it.SourcePath)
		infos[i].segments = segs
		if len(segs) == 0 {
			continue
		}
		root := strings.ToLower(segs[0])
		packCount[root]++
		if len(segs) >= 2 {
			packHasNested[root] = true
		}
	}

	for i, info := range infos {
		if len(info.segments) == 0 {
			continue
		}
		root := strings.ToLower(info.segments[0])
		if packHasNested[root] && packCount[root] >= 2 {
			families[i] = root
		}
	}
}

func overlaySourcePkgPacks(items []domain.Item, families []string) {
	pkgCount := make(map[string]int)
	pkgFamily := make(map[string]string)
	for _, it := range items {
		key := normalizeSourcePkg(it.SourcePkg)
		if key == "" {
			continue
		}
		pkgCount[key]++
		if _, ok := pkgFamily[key]; !ok {
			pkgFamily[key] = sourcePkgSlug(key)
		}
	}

	for i, it := range items {
		key := normalizeSourcePkg(it.SourcePkg)
		if key == "" || pkgCount[key] < 2 {
			continue
		}
		families[i] = pkgFamily[key]
	}
}

func overlayAliasPacks(items []domain.Item, families []string) {
	names := make(map[string]bool)
	for _, it := range items {
		n := strings.ToLower(strings.TrimSpace(it.Name))
		if n != "" {
			names[n] = true
		}
	}

	for _, alias := range loadedPackAliases {
		parent := strings.ToLower(strings.TrimSpace(alias.Parent))
		family := strings.ToLower(strings.TrimSpace(alias.Family))
		if parent == "" || family == "" || !names[parent] {
			continue
		}

		var regexes []interface{ MatchString(string) bool }
		for _, pat := range alias.Patterns {
			re := patternToRegex(pat)
			if re != nil {
				regexes = append(regexes, re)
			}
		}

		for i, it := range items {
			name := strings.TrimSpace(it.Name)
			if name == "" {
				continue
			}
			for _, re := range regexes {
				if re.MatchString(name) {
					families[i] = family
					break
				}
			}
		}
	}
}

func stampParents(items []domain.Item, families []string) []domain.Item {
	familyChildren := make(map[string]map[string]bool)
	for i, fam := range families {
		if fam == "" {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(items[i].Name))
		if name == "" || isPackParentName(name, fam) || isDirectoryPackRoot(items[i], fam) {
			continue
		}
		if familyChildren[fam] == nil {
			familyChildren[fam] = make(map[string]bool)
		}
		familyChildren[fam][name] = true
	}

	for i := range items {
		fam := families[i]
		if fam == "" {
			items[i].Family = ""
			items[i].IsParent = false
			items[i].ChildCount = 0
			continue
		}

		items[i].Family = fam
		name := strings.ToLower(strings.TrimSpace(items[i].Name))
		childCount := len(familyChildren[fam])
		if (isPackParentName(name, fam) || isDirectoryPackRoot(items[i], fam)) && childCount > 0 {
			items[i].IsParent = true
			items[i].ChildCount = childCount
		} else {
			items[i].IsParent = false
			items[i].ChildCount = 0
		}
	}

	return items
}

func isDirectoryPackRoot(it domain.Item, family string) bool {
	segs := skillsRelSegments(it.SourcePath)
	return len(segs) == 1 && strings.EqualFold(segs[0], family)
}

func isPackParentName(name, family string) bool {
	if name == "" || family == "" {
		return false
	}
	if name == family {
		return true
	}
	trimmed := strings.TrimSuffix(family, "-skills")
	trimmed = strings.TrimSuffix(trimmed, "-skill")
	return name == trimmed && trimmed != family
}

func normalizeSourcePkg(sourcePkg string) string {
	s := strings.TrimSpace(strings.ToLower(sourcePkg))
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimPrefix(s, "https://github.com/")
	s = strings.TrimPrefix(s, "http://github.com/")
	s = strings.TrimPrefix(s, "git@github.com:")
	return strings.TrimSuffix(s, ".git")
}

func sourcePkgSlug(normalizedPkg string) string {
	if normalizedPkg == "" {
		return ""
	}
	if i := strings.LastIndex(normalizedPkg, "/"); i >= 0 && i < len(normalizedPkg)-1 {
		return normalizedPkg[i+1:]
	}
	return normalizedPkg
}

func skillsRelSegments(sourcePath string) []string {
	if sourcePath == "" {
		return nil
	}
	p := filepath.ToSlash(sourcePath)
	if strings.EqualFold(filepath.Base(p), "SKILL.md") {
		p = filepath.ToSlash(filepath.Dir(sourcePath))
	}
	lower := strings.ToLower(p)
	marker := "/skills/"
	idx := strings.LastIndex(lower, marker)
	if idx < 0 {
		return nil
	}
	rest := strings.Trim(p[idx+len(marker):], "/")
	if rest == "" {
		return nil
	}
	return strings.Split(rest, "/")
}
