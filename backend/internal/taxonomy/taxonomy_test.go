package taxonomy

import (
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestEngineClassification(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name            string
		item            domain.Item
		wantCategory    string
		wantSubCategory string
		wantClassified  bool
	}{
		{
			name: "Hunt SQLi Web Vulnerability",
			item: domain.Item{
				ID:          "cursor:skill:hunt-sqli",
				Name:        "hunt-sqli",
				Description: "Detect SQL injection flaws in endpoints",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Web Vulnerabilities",
			wantClassified:  true,
		},
		{
			name: "XSS by pattern without false positive",
			item: domain.Item{
				ID:          "global:skill:xss-payloads",
				Name:        "xss-payloads",
				Description: "Generate cross-site scripting attack vectors",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Web Vulnerabilities",
			wantClassified:  true,
		},
		{
			name: "Recon & OSINT with bb-* pattern",
			item: domain.Item{
				ID:          "global:skill:bb-local-toolkit",
				Name:        "bb-local-toolkit",
				Description: "Bug bounty reconnaissance toolkit",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Recon & OSINT",
			wantClassified:  true,
		},
		{
			name: "Red Team & Exploits with apk-redteam-*",
			item: domain.Item{
				ID:          "global:skill:apk-redteam-pipeline",
				Name:        "apk-redteam-pipeline",
				Description: "Android APK reverse engineering and red team attack path",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Red Team & Exploits",
			wantClassified:  true,
		},
		{
			name: "Audit & Compliance with *-audit pattern",
			item: domain.Item{
				ID:          "global:skill:security-audit",
				Name:        "security-audit",
				Description: "Performs dependency vulnerability audit",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Audit & Compliance",
			wantClassified:  true,
		},
		{
			name: "Gentle AI SDD with sdd-*",
			item: domain.Item{
				ID:          "global:skill:sdd-orchestrator",
				Name:        "sdd-orchestrator",
				Description: "Spec-Driven Development orchestrator agent",
			},
			wantCategory:    "Gentle AI & SDD",
			wantSubCategory: "Spec-Driven Development",
			wantClassified:  true,
		},
		{
			name: "Gentle AI Review with gentle-ai-*",
			item: domain.Item{
				ID:          "global:skill:gentle-ai-bench",
				Name:        "gentle-ai-bench",
				Description: "Benchmark framework for Gentle AI agents",
			},
			wantCategory:    "Gentle AI & SDD",
			wantSubCategory: "Review & Benchmark",
			wantClassified:  true,
		},
		{
			name: "Frontend Aesthetics with brutalist-*",
			item: domain.Item{
				ID:          "global:skill:industrial-brutalist-ui",
				Name:        "industrial-brutalist-ui",
				Description: "Generate clean brutalist interfaces",
			},
			wantCategory:    "Frontend & UI/UX",
			wantSubCategory: "Aesthetics & Taste",
			wantClassified:  true,
		},
		{
			name: "Frontend Design Systems with brandkit",
			item: domain.Item{
				ID:          "global:skill:brandkit",
				Name:        "brandkit",
				Description: "Export and enforce design tokens",
			},
			wantCategory:    "Frontend & UI/UX",
			wantSubCategory: "Design Systems & Tokens",
			wantClassified:  true,
		},
		{
			name: "Visual AI with baoyu-*",
			item: domain.Item{
				ID:          "global:skill:baoyu-infographics",
				Name:        "baoyu-infographics",
				Description: "Infographics and visual diagrams",
			},
			wantCategory:    "Frontend & UI/UX",
			wantSubCategory: "Visual AI & Image-to-Code",
			wantClassified:  true,
		},
		{
			name: "Workflow Git Automation with *-pr",
			item: domain.Item{
				ID:          "global:skill:branch-pr",
				Name:        "branch-pr",
				Description: "Create branches and PRs automatically",
			},
			wantCategory:    "Workflow & Engineering",
			wantSubCategory: "Git & PR Automation",
			wantClassified:  true,
		},
		{
			name: "Skill Management with find-skills",
			item: domain.Item{
				ID:          "global:skill:find-skills",
				Name:        "find-skills",
				Description: "Search skills in ecosystem",
			},
			wantCategory:    "AI Core & Meta",
			wantSubCategory: "Skill Management",
			wantClassified:  true,
		},
		{
			name: "Cloud & Integrations with trello",
			item: domain.Item{
				ID:          "hermes:mcp:trello",
				Name:        "trello",
				Description: "Manage boards and cards in Trello",
			},
			wantCategory:    "Cloud & Integrations",
			wantSubCategory: "Productivity & Services",
			wantClassified:  true,
		},
		{
			name: "Supply Chain Attack Recon should be Bug Bounty Recon not Frontend Bundler",
			item: domain.Item{
				ID:          "global:skill:supply-chain-attack-recon",
				Name:        "supply-chain-attack-recon",
				Description: "Reconnaissance for package squatting, dependency confusion when finding internal package names in JS bundles.",
			},
			wantCategory:    "Bug Bounty & Security",
			wantSubCategory: "Recon & OSINT",
			wantClassified:  true,
		},
		{
			name: "Unclassified item pending triage",
			item: domain.Item{
				ID:          "custom:unknown:xyz-unheard-tool",
				Name:        "xyz-unheard-tool",
				Description: "Some completely random unknown tool with no matching keywords",
			},
			wantCategory:    "Unclassified",
			wantSubCategory: "Pending Triage",
			wantClassified:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, subCat, isClassified := engine.Classify(tt.item)
			if isClassified != tt.wantClassified {
				t.Errorf("Classify(%s) isClassified = %v, want %v", tt.item.Name, isClassified, tt.wantClassified)
			}
			if cat != tt.wantCategory {
				t.Errorf("Classify(%s) Category = %q, want %q", tt.item.Name, cat, tt.wantCategory)
			}
			if subCat != tt.wantSubCategory {
				t.Errorf("Classify(%s) SubCategory = %q, want %q", tt.item.Name, subCat, tt.wantSubCategory)
			}
		})
	}
}

func TestAddRuleAndPersist(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "taxonomy.json")

	// Write initial simple taxonomy
	initialData := []byte(`{
		"rules": [
			{
				"category": "Custom Category",
				"subCategory": "Custom SubCategory",
				"patterns": ["custom-tool-*"]
			}
		]
	}`)
	if err := os.WriteFile(tempFile, initialData, 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := NewEngineWithFile(tempFile)
	if err != nil {
		t.Fatalf("failed to create engine with file: %v", err)
	}

	// Test unclassified before rule added
	testItem := domain.Item{
		ID:   "custom:tool:foobar-scanner",
		Name: "foobar-scanner",
	}
	_, _, classified := engine.Classify(testItem)
	if classified {
		t.Fatalf("expected foobar-scanner to be unclassified initially")
	}

	// Add dynamic rule
	err = engine.AddRule("DevSecOps", "Vulnerability Scanning", []string{"foobar-*"})
	if err != nil {
		t.Fatalf("AddRule failed: %v", err)
	}

	// Now should be classified
	cat, subCat, classified := engine.Classify(testItem)
	if !classified {
		t.Fatalf("expected foobar-scanner to be classified after AddRule")
	}
	if cat != "DevSecOps" || subCat != "Vulnerability Scanning" {
		t.Errorf("got (%s, %s), want (DevSecOps, Vulnerability Scanning)", cat, subCat)
	}

	// Re-load engine from same file to verify persistence
	reloaded, err := NewEngineWithFile(tempFile)
	if err != nil {
		t.Fatalf("failed to reload engine: %v", err)
	}
	cat2, subCat2, classified2 := reloaded.Classify(testItem)
	if !classified2 || cat2 != "DevSecOps" || subCat2 != "Vulnerability Scanning" {
		t.Errorf("persisted reload failed: got (%s, %s, %v)", cat2, subCat2, classified2)
	}
}
