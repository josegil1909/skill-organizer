package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestExtractFrontmatterAndBody(t *testing.T) {
	content := `---
name: my-skill
description: A helpful skill description.
---

# My Skill Header

This is the first paragraph of the skill body.
It has multiple lines of text to test preview extraction.
`
	fm, preview, rawConfig := extractFrontmatterAndBody(content)
	if fm.Name != "my-skill" {
		t.Fatalf("expected name 'my-skill', got '%s'", fm.Name)
	}
	if fm.Description != "A helpful skill description." {
		t.Fatalf("expected description 'A helpful skill description.', got '%s'", fm.Description)
	}
	if preview == "" {
		t.Fatal("expected non-empty body preview")
	}
	if rawConfig == "" {
		t.Fatal("expected non-empty rawConfig")
	}
}

func TestSkillsScannerDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create skill 1 with frontmatter
	skill1Dir := filepath.Join(tempDir, "skill1")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatal(err)
	}
	skill1File := filepath.Join(skill1Dir, "SKILL.md")
	if err := os.WriteFile(skill1File, []byte("---\nname: skill-one\ndescription: First test skill\n---\nBody here"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create nested skill 2 without frontmatter
	skill2Dir := filepath.Join(tempDir, "category", "skill2")
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatal(err)
	}
	skill2File := filepath.Join(skill2Dir, "SKILL.md")
	if err := os.WriteFile(skill2File, []byte("# Header\nDirect description without frontmatter"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create hidden directory with skill that must be ignored
	hiddenDir := filepath.Join(tempDir, ".archive", "ignored")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "SKILL.md"), []byte("---\nname: ignored\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(Target{Path: tempDir, Provider: domain.ProviderCursor})
	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	var foundSkill1, foundSkill2 bool
	for _, it := range items {
		if it.Name == "skill-one" {
			foundSkill1 = true
			if it.Provider != domain.ProviderCursor {
				t.Errorf("expected provider cursor, got %s", it.Provider)
			}
		}
		if it.Name == "skill2" { // derived from parent dir name
			foundSkill2 = true
		}
		if it.Name == "ignored" {
			t.Error("hidden directory skill should have been ignored")
		}
	}

	if !foundSkill1 || !foundSkill2 {
		t.Fatalf("expected skill-one and skill2, got foundSkill1=%v, foundSkill2=%v", foundSkill1, foundSkill2)
	}
}

func TestSkillOriginClassification(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Hermes manifest setup
	hermesManifestPath := filepath.Join(tempDir, ".bundled_manifest")
	if err := os.WriteFile(hermesManifestPath, []byte("airtable:30f47a4b\napple-notes:5e448abf\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Skill lock setup
	skillLockPath := filepath.Join(tempDir, ".skill-lock.json")
	lockJSON := `{
		"skills": {
			"security-audit": {
				"source": "cloudflare/security-audit-skill",
				"sourceUrl": "https://github.com/cloudflare/security-audit-skill.git"
			}
		}
	}`
	if err := os.WriteFile(skillLockPath, []byte(lockJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Hermes skills dir: one bundled ("airtable"), one custom ("my-hermes-skill")
	hermesSkillsDir := filepath.Join(tempDir, "hermes-skills")
	airtableDir := filepath.Join(hermesSkillsDir, "airtable")
	_ = os.MkdirAll(airtableDir, 0755)
	_ = os.WriteFile(filepath.Join(airtableDir, "SKILL.md"), []byte("# Airtable"), 0644)

	customHermesDir := filepath.Join(hermesSkillsDir, "my-hermes-skill")
	_ = os.MkdirAll(customHermesDir, 0755)
	_ = os.WriteFile(filepath.Join(customHermesDir, "SKILL.md"), []byte("# Custom Hermes"), 0644)

	// Create Global skills dir: one locked ("security-audit"), one custom ("my-global-skill")
	globalSkillsDir := filepath.Join(tempDir, "global-skills")
	secAuditDir := filepath.Join(globalSkillsDir, "security-audit")
	_ = os.MkdirAll(secAuditDir, 0755)
	_ = os.WriteFile(filepath.Join(secAuditDir, "SKILL.md"), []byte("# Security Audit"), 0644)

	customGlobalDir := filepath.Join(globalSkillsDir, "my-global-skill")
	_ = os.MkdirAll(customGlobalDir, 0755)
	_ = os.WriteFile(filepath.Join(customGlobalDir, "SKILL.md"), []byte("# Custom Global"), 0644)

	// Create Grok skills dirs: bundled/skills vs regular grok skills
	grokBundledDir := filepath.Join(tempDir, "bundled", "skills", "build-with-ai")
	_ = os.MkdirAll(grokBundledDir, 0755)
	_ = os.WriteFile(filepath.Join(grokBundledDir, "SKILL.md"), []byte("# Build With AI"), 0644)

	grokCustomDir := filepath.Join(tempDir, "grok-user-skills", "trello")
	_ = os.MkdirAll(grokCustomDir, 0755)
	_ = os.WriteFile(filepath.Join(grokCustomDir, "SKILL.md"), []byte("# Trello"), 0644)

	scanner := NewScanner(
		Target{Path: hermesSkillsDir, Provider: domain.ProviderHermes},
		Target{Path: globalSkillsDir, Provider: domain.ProviderGlobal},
		Target{Path: filepath.Join(tempDir, "bundled", "skills"), Provider: domain.ProviderGrok},
		Target{Path: grokCustomDir, Provider: domain.ProviderGrok},
	).WithManifestPaths(hermesManifestPath, skillLockPath)

	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}

	itemMap := make(map[string]domain.Item)
	for _, it := range items {
		itemMap[it.Name] = it
	}

	// 1. Hermes airtable -> builtin
	if it, ok := itemMap["airtable"]; !ok || it.Origin != domain.OriginBuiltin {
		t.Errorf("expected airtable origin 'builtin', got %v", it.Origin)
	}
	// 2. Hermes my-hermes-skill -> custom
	if it, ok := itemMap["my-hermes-skill"]; !ok || it.Origin != domain.OriginCustom {
		t.Errorf("expected my-hermes-skill origin 'custom', got %v", it.Origin)
	}
	// 3. Global security-audit -> installed with SourceURL
	if it, ok := itemMap["security-audit"]; !ok {
		t.Error("expected security-audit item")
	} else {
		if it.Origin != domain.OriginInstalled {
			t.Errorf("expected security-audit origin 'installed', got %s", it.Origin)
		}
		if it.SourceURL != "https://github.com/cloudflare/security-audit-skill.git" {
			t.Errorf("expected sourceUrl 'https://github.com/cloudflare/security-audit-skill.git', got %s", it.SourceURL)
		}
		if it.SourcePkg != "cloudflare/security-audit-skill" {
			t.Errorf("expected sourcePkg 'cloudflare/security-audit-skill', got %s", it.SourcePkg)
		}
		if it.RegistryURL != "https://skills.sh/cloudflare/security-audit-skill/security-audit" {
			t.Errorf("expected registryUrl 'https://skills.sh/cloudflare/security-audit-skill/security-audit', got %s", it.RegistryURL)
		}
	}
	// 4. Global my-global-skill -> custom
	if it, ok := itemMap["my-global-skill"]; !ok || it.Origin != domain.OriginCustom {
		t.Errorf("expected my-global-skill origin 'custom', got %v", it.Origin)
	}
	// 5. Grok bundled -> builtin
	if it, ok := itemMap["build-with-ai"]; !ok || it.Origin != domain.OriginBuiltin {
		t.Errorf("expected build-with-ai origin 'builtin', got %v", it.Origin)
	}
	// 6. Grok custom -> custom
	if it, ok := itemMap["trello"]; !ok || it.Origin != domain.OriginCustom {
		t.Errorf("expected trello origin 'custom', got %v", it.Origin)
	}
}

