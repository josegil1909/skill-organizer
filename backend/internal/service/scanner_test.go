package service

import (
	"context"
	"testing"

	"organizer/backend/internal/domain"
	"organizer/backend/internal/taxonomy"
)

type mockAdapter struct {
	name  string
	items []domain.Item
	err   error
}

func (m *mockAdapter) Name() string {
	return m.name
}

func (m *mockAdapter) Scan(ctx context.Context) ([]domain.Item, error) {
	return m.items, m.err
}

func TestAggregatorService(t *testing.T) {
	mock1 := &mockAdapter{
		name: "mock1",
		items: []domain.Item{
			{
				ID:          "cursor:mcp:unknown-mcp",
				Name:        "unknown-mcp",
				Type:        domain.ItemTypeMCP,
				Provider:    domain.ProviderCursor,
				Origin:      domain.OriginInstalled,
				Description: "Unknown MCP analyzer",
			},
			{
				ID:          "global:skill:security-audit",
				Name:        "security-audit",
				Type:        domain.ItemTypeSkill,
				Provider:    domain.ProviderGlobal,
				Origin:      domain.OriginCustom,
				Description: "Security audit tool",
			},
		},
	}

	mock2 := &mockAdapter{
		name: "mock2",
		items: []domain.Item{
			{
				ID:          "cursor:mcp:unknown-mcp", // Duplicate ID should be deduplicated
				Name:        "unknown-mcp",
				Type:        domain.ItemTypeMCP,
				Provider:    domain.ProviderCursor,
				Origin:      domain.OriginInstalled,
				Description: "Duplicate",
			},
			{
				ID:          "claude:mcp:unknown-mcp",
				Name:        "unknown-mcp",
				Type:        domain.ItemTypeMCP,
				Provider:    domain.ProviderClaude,
				Origin:      domain.OriginBuiltin,
				Description: "Unknown MCP for claude",
			},
		},
	}

	svc := NewAggregatorService(mock1, mock2)
	svc.SetTaxonomy(taxonomy.NewDefaultEngine())
	scanned, err := svc.ScanAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scanned) != 3 {
		t.Fatalf("expected 3 deduplicated items, got %d", len(scanned))
	}

	// Test Filter by Type
	mcpItems := svc.GetItems(domain.ItemTypeMCP, "", "", "", "", "", nil, "")
	if len(mcpItems) != 2 {
		t.Fatalf("expected 2 MCP items, got %d", len(mcpItems))
	}

	// Test Filter by Provider
	globalItems := svc.GetItems("", domain.ProviderGlobal, "", "", "", "", nil, "")
	if len(globalItems) != 1 {
		t.Fatalf("expected 1 global item, got %d", len(globalItems))
	}

	// Test Filter by Origin
	customItems := svc.GetItems("", "", domain.OriginCustom, "", "", "", nil, "")
	if len(customItems) != 1 || customItems[0].Name != "security-audit" {
		t.Fatalf("expected 1 custom item 'security-audit', got %v", customItems)
	}

	installedItems := svc.GetItems("", "", domain.OriginInstalled, "", "", "", nil, "")
	if len(installedItems) != 1 || installedItems[0].Name != "unknown-mcp" {
		t.Fatalf("expected 1 installed item 'unknown-mcp', got %v", installedItems)
	}

	builtinItems := svc.GetItems("", "", domain.OriginBuiltin, "", "", "", nil, "")
	if len(builtinItems) != 1 || builtinItems[0].Provider != domain.ProviderClaude {
		t.Fatalf("expected 1 builtin item from claude, got %v", builtinItems)
	}

	// Test Search
	searchAudit := svc.GetItems("", "", "", "", "", "", nil, "Security")
	if len(searchAudit) != 1 || searchAudit[0].Name != "security-audit" {
		t.Fatalf("expected audit item for search 'Security', got %v", searchAudit)
	}

	// Test Classification Filtering: security-audit should be classified under "Bug Bounty & Security"
	classifiedTrue := true
	classifiedFalse := false
	classItems := svc.GetItems("", "", "", "", "", "", &classifiedTrue, "")
	if len(classItems) != 1 || classItems[0].Category != "Bug Bounty & Security" {
		t.Fatalf("expected 1 classified item, got %v", classItems)
	}

	unclassItems := svc.GetItems("", "", "", "", "", "", &classifiedFalse, "")
	if len(unclassItems) != 2 {
		t.Fatalf("expected 2 unclassified items, got %d", len(unclassItems))
	}

	// Test Category Filtering
	secItems := svc.GetItems("", "", "", "Bug Bounty & Security", "", "", nil, "")
	if len(secItems) != 1 || secItems[0].Name != "security-audit" {
		t.Fatalf("expected 1 security item, got %v", secItems)
	}

	// Test SubCategory Filtering
	auditSubItems := svc.GetItems("", "", "", "Bug Bounty & Security", "Audit & Compliance", "", nil, "")
	if len(auditSubItems) != 1 || auditSubItems[0].Name != "security-audit" {
		t.Fatalf("expected 1 audit & compliance item, got %v", auditSubItems)
	}

	// Test Stats
	stats := svc.GetStats()
	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
	if stats.ByType[domain.ItemTypeMCP] != 2 {
		t.Errorf("expected 2 MCPs, got %d", stats.ByType[domain.ItemTypeMCP])
	}
	if stats.ByType[domain.ItemTypeSkill] != 1 {
		t.Errorf("expected 1 Skill, got %d", stats.ByType[domain.ItemTypeSkill])
	}
	if stats.ClassifiedCount != 1 {
		t.Errorf("expected ClassifiedCount 1, got %d", stats.ClassifiedCount)
	}
	if stats.UnclassifiedCount != 2 {
		t.Errorf("expected UnclassifiedCount 2, got %d", stats.UnclassifiedCount)
	}
	if stats.ByCategory["Bug Bounty & Security"] != 1 {
		t.Errorf("expected 1 Bug Bounty & Security in ByCategory, got %d", stats.ByCategory["Bug Bounty & Security"])
	}
}

func TestFamilyFilteringAndStats(t *testing.T) {
	svc := NewAggregatorService()
	svc.SetItems([]domain.Item{
		{ID: "1", Name: "vue", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "vue-best-practices", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "vue-debug-guides", Type: domain.ItemTypeSkill},
		{ID: "4", Name: "hunt-xss", Type: domain.ItemTypeSkill},
		{ID: "5", Name: "hunt-sqli", Type: domain.ItemTypeSkill},
	})

	stats := svc.GetStats()
	if stats.ByFamily["vue"] != 3 {
		t.Errorf("expected ByFamily['vue'] == 3, got %d", stats.ByFamily["vue"])
	}
	if stats.ByFamily["hunt"] != 2 {
		t.Errorf("expected ByFamily['hunt'] == 2, got %d", stats.ByFamily["hunt"])
	}

	// Test GetItems with family filter
	vueItems := svc.GetItems("", "", "", "", "", "vue", nil, "")
	if len(vueItems) != 3 {
		t.Errorf("expected 3 vue family items, got %d", len(vueItems))
	}

	huntItems := svc.GetItems("", "", "", "", "", "hunt", nil, "")
	if len(huntItems) != 2 {
		t.Errorf("expected 2 hunt family items, got %d", len(huntItems))
	}
}
