package taxonomy

import (
	"testing"

	"organizer/backend/internal/domain"
)

func TestInferFamilies_ParentWithChildren(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "vue", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "vue-best-practices", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "vue-debug-guides", Type: domain.ItemTypeSkill},
		{ID: "4", Name: "vue-pinia-best-practices", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)

	// Item "vue" should be the parent
	var parent *domain.Item
	var children []domain.Item

	for i := range result {
		if result[i].Name == "vue" {
			parent = &result[i]
		} else {
			children = append(children, result[i])
		}
	}

	if parent == nil {
		t.Fatalf("expected 'vue' to exist in result")
	}
	if parent.Family != "vue" {
		t.Errorf("expected parent family to be 'vue', got '%s'", parent.Family)
	}
	if !parent.IsParent {
		t.Errorf("expected parent IsParent to be true")
	}
	if parent.ChildCount != 3 {
		t.Errorf("expected parent ChildCount to be 3, got %d", parent.ChildCount)
	}

	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	for _, ch := range children {
		if ch.Family != "vue" {
			t.Errorf("expected child %s to have family 'vue', got '%s'", ch.Name, ch.Family)
		}
		if ch.IsParent {
			t.Errorf("expected child %s IsParent to be false", ch.Name)
		}
		if ch.ChildCount != 0 {
			t.Errorf("expected child %s ChildCount to be 0, got %d", ch.Name, ch.ChildCount)
		}
	}
}

func TestInferFamilies_NoStandaloneParent(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "hunt-xss", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "hunt-sqli", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "hunt-idor", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)

	for _, it := range result {
		if it.Family != "hunt" {
			t.Errorf("expected item %s to have family 'hunt', got '%s'", it.Name, it.Family)
		}
		if it.IsParent {
			t.Errorf("expected item %s IsParent to be false when no standalone item exists", it.Name)
		}
		if it.ChildCount != 0 {
			t.Errorf("expected item %s ChildCount to be 0, got %d", it.Name, it.ChildCount)
		}
	}
}

func TestInferFamilies_CommonFamilies(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "sdd-apply", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "sdd-verify", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "game-tilesets", Type: domain.ItemTypeSkill},
		{ID: "4", Name: "game-ui-icons", Type: domain.ItemTypeSkill},
		{ID: "5", Name: "golang-patterns", Type: domain.ItemTypeSkill},
		{ID: "6", Name: "golang-testing", Type: domain.ItemTypeSkill},
		{ID: "7", Name: "bb-methodology", Type: domain.ItemTypeSkill},
		{ID: "8", Name: "recon-subdomains", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)

	familyMap := make(map[string]string)
	for _, it := range result {
		familyMap[it.Name] = it.Family
	}

	expected := map[string]string{
		"sdd-apply":        "sdd",
		"sdd-verify":       "sdd",
		"game-tilesets":    "game",
		"game-ui-icons":    "game",
		"golang-patterns":  "golang",
		"golang-testing":   "golang",
		"bb-methodology":   "bb",
		"recon-subdomains": "recon",
	}

	for name, expFam := range expected {
		if familyMap[name] != expFam {
			t.Errorf("for item %s expected family %s, got %s", name, expFam, familyMap[name])
		}
	}
}

func TestInferFamilies_MultiSegmentPrefix(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "image-to-code", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "image-to-code-generator", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)

	if result[0].Family != "image-to-code" {
		t.Errorf("expected family 'image-to-code', got '%s'", result[0].Family)
	}
	if !result[0].IsParent {
		t.Errorf("expected 'image-to-code' to be parent")
	}
	if result[0].ChildCount != 1 {
		t.Errorf("expected 'image-to-code' ChildCount to be 1, got %d", result[0].ChildCount)
	}

	if result[1].Family != "image-to-code" {
		t.Errorf("expected family 'image-to-code', got '%s'", result[1].Family)
	}
	if result[1].IsParent {
		t.Errorf("expected child IsParent to be false")
	}
}

func TestInferFamilies_BugBountyPackGroupsHuntWhenParentPresent(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "bug-bounty", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "hunt-xss", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "hunt-sqli", Type: domain.ItemTypeSkill},
		{ID: "4", Name: "bb-methodology", Type: domain.ItemTypeSkill},
		{ID: "5", Name: "report-writing", Type: domain.ItemTypeSkill},
		{ID: "6", Name: "vue", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)
	got := map[string]domain.Item{}
	for _, it := range result {
		got[it.Name] = it
	}

	parent := got["bug-bounty"]
	if parent.Family != "bug-bounty" {
		t.Errorf("parent family = %q, want bug-bounty", parent.Family)
	}
	if !parent.IsParent {
		t.Errorf("expected bug-bounty IsParent")
	}
	if parent.ChildCount != 4 {
		t.Errorf("bug-bounty ChildCount = %d, want 4", parent.ChildCount)
	}

	for _, name := range []string{"hunt-xss", "hunt-sqli", "bb-methodology", "report-writing"} {
		it := got[name]
		if it.Family != "bug-bounty" {
			t.Errorf("%s family = %q, want bug-bounty", name, it.Family)
		}
		if it.IsParent {
			t.Errorf("%s should not be parent", name)
		}
	}

	if got["vue"].Family != "vue" && got["vue"].Family != "" {
		// vue alone is a common family prefix with no children
		if got["vue"].Family != "" {
			t.Errorf("vue should not join bug-bounty pack, got family %q", got["vue"].Family)
		}
	}
	if got["vue"].Family == "bug-bounty" {
		t.Errorf("vue must not join bug-bounty pack")
	}
}

func TestInferFamilies_AliasDoesNotApplyWithoutParent(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "hunt-xss", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "report-writing", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)
	for _, it := range result {
		if it.Family == "bug-bounty" {
			t.Errorf("%s got family bug-bounty without parent skill present", it.Name)
		}
	}
}

func TestInferFamilies_SourcePkgGroupsRepoSiblings(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "brandkit", Type: domain.ItemTypeSkill, SourcePkg: "leonxlnx/taste-skill"},
		{ID: "2", Name: "gpt-taste", Type: domain.ItemTypeSkill, SourcePkg: "leonxlnx/taste-skill"},
		{ID: "3", Name: "impeccable", Type: domain.ItemTypeSkill, SourcePkg: "leonxlnx/taste-skill"},
		{ID: "4", Name: "pdf", Type: domain.ItemTypeSkill, SourcePkg: "anthropics/skills"},
	}

	result := InferFamilies(items)
	got := map[string]domain.Item{}
	for _, it := range result {
		got[it.Name] = it
	}

	for _, name := range []string{"brandkit", "gpt-taste", "impeccable"} {
		if got[name].Family != "taste-skill" {
			t.Errorf("%s family = %q, want taste-skill", name, got[name].Family)
		}
	}
	if got["pdf"].Family == "taste-skill" {
		t.Errorf("pdf must not join taste-skill pack")
	}
}

func TestInferFamilies_SourcePkgMarksRepoSlugAsParent(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "hack", Type: domain.ItemTypeSkill, SourcePkg: "yaklang/hack-skills"},
		{ID: "2", Name: "sqli-sql-injection", Type: domain.ItemTypeSkill, SourcePkg: "yaklang/hack-skills"},
		{ID: "3", Name: "xss-cross-site-scripting", Type: domain.ItemTypeSkill, SourcePkg: "yaklang/hack-skills"},
	}

	result := InferFamilies(items)
	got := map[string]domain.Item{}
	for _, it := range result {
		got[it.Name] = it
	}

	if got["hack"].Family != "hack-skills" {
		t.Errorf("hack family = %q, want hack-skills", got["hack"].Family)
	}
	if !got["hack"].IsParent {
		t.Errorf("expected repo master entry 'hack' to be parent")
	}
	if got["hack"].ChildCount != 2 {
		t.Errorf("hack ChildCount = %d, want 2", got["hack"].ChildCount)
	}
	if got["sqli-sql-injection"].Family != "hack-skills" || got["sqli-sql-injection"].IsParent {
		t.Errorf("sqli-sql-injection should be a child of hack-skills")
	}
}

func TestInferFamilies_DirectoryPackFromNestedSkillsPath(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "erp", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/SKILL.md"},
		{ID: "2", Name: "vue", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/vue/SKILL.md"},
		{ID: "3", Name: "vite", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/vite/SKILL.md"},
		{ID: "4", Name: "hunt-xss", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.agents/skills/hunt-xss/SKILL.md"},
	}

	result := InferFamilies(items)
	got := map[string]domain.Item{}
	for _, it := range result {
		got[it.Name] = it
	}

	if got["erp"].Family != "erp" {
		t.Errorf("erp family = %q, want erp", got["erp"].Family)
	}
	if !got["erp"].IsParent {
		t.Errorf("expected erp SKILL.md to be directory pack parent")
	}
	if got["vue"].Family != "erp" {
		t.Errorf("vue family = %q, want erp (directory pack beats vue prefix)", got["vue"].Family)
	}
	if got["vite"].Family != "erp" {
		t.Errorf("vite family = %q, want erp", got["vite"].Family)
	}
	if got["hunt-xss"].Family == "erp" {
		t.Errorf("flat hunt-xss must not join erp directory pack")
	}
}

func TestInferFamilies_DirectoryPackParentUsesFolderEvenIfRenamed(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "ERP suite", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/SKILL.md"},
		{ID: "2", Name: "vue", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/vue/SKILL.md"},
		{ID: "3", Name: "vite", Type: domain.ItemTypeSkill, SourcePath: "/home/u/.hermes/skills/erp/vite/SKILL.md"},
	}

	result := InferFamilies(items)
	var parent *domain.Item
	for i := range result {
		if result[i].Name == "ERP suite" {
			parent = &result[i]
		}
	}
	if parent == nil {
		t.Fatal("expected ERP suite item")
	}
	if parent.Family != "erp" {
		t.Errorf("family = %q, want erp", parent.Family)
	}
	if !parent.IsParent {
		t.Errorf("expected folder-root SKILL.md to be parent even when display name differs")
	}
	if parent.ChildCount != 2 {
		t.Errorf("ChildCount = %d, want 2", parent.ChildCount)
	}
}

func TestInferFamilies_DynamicPrefixPair(t *testing.T) {
	items := []domain.Item{
		{ID: "1", Name: "mytesttool-alpha", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "mytesttool-beta", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "single-isolated-item", Type: domain.ItemTypeSkill},
	}

	result := InferFamilies(items)

	if result[0].Family != "mytesttool" {
		t.Errorf("expected dynamic family 'mytesttool', got '%s'", result[0].Family)
	}
	if result[1].Family != "mytesttool" {
		t.Errorf("expected dynamic family 'mytesttool', got '%s'", result[1].Family)
	}

	// Single isolated item not in common families should have empty family
	if result[2].Family != "" {
		t.Errorf("expected single isolated item to have empty family, got '%s'", result[2].Family)
	}
}
