package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
	"organizer/backend/internal/service"
	"organizer/backend/internal/taxonomy"
)

func setupTestServer() (*Server, *service.AggregatorService) {
	svc := service.NewAggregatorService()
	svc.SetTaxonomy(taxonomy.NewDefaultEngine())
	svc.SetItems([]domain.Item{
		{
			ID:          "cursor:mcp:unknown-mcp",
			Name:        "unknown-mcp",
			Type:        domain.ItemTypeMCP,
			Provider:    domain.ProviderCursor,
			Origin:      domain.OriginInstalled,
			Description: "Unknown MCP tool",
		},
		{
			ID:          "global:skill:audit",
			Name:        "security-audit",
			Type:        domain.ItemTypeSkill,
			Provider:    domain.ProviderGlobal,
			Origin:      domain.OriginCustom,
			Description: "Audit security tool",
		},
	})
	srv := NewServer(svc)
	return srv, svc
}

func TestHandleHealth(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", res["status"])
	}
}

func TestHandleGetItems(t *testing.T) {
	srv, _ := setupTestServer()

	// 1. Get all items
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var items []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// 2. Filter by type
	req = httptest.NewRequest(http.MethodGet, "/api/items?type=skill", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var filtered []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&filtered); err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].Name != "security-audit" {
		t.Fatalf("expected only security-audit, got %v", filtered)
	}

	// 3. Filter by origin
	req = httptest.NewRequest(http.MethodGet, "/api/items?origin=installed", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var originFiltered []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&originFiltered); err != nil {
		t.Fatal(err)
	}
	if len(originFiltered) != 1 || originFiltered[0].Name != "unknown-mcp" {
		t.Fatalf("expected only unknown-mcp, got %v", originFiltered)
	}

	// 4. Filter by isClassified
	req = httptest.NewRequest(http.MethodGet, "/api/items?isClassified=true", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var classFiltered []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&classFiltered); err != nil {
		t.Fatal(err)
	}
	if len(classFiltered) != 1 || classFiltered[0].Name != "security-audit" {
		t.Fatalf("expected only classified security-audit, got %v", classFiltered)
	}

	// 5. Filter by category
	req = httptest.NewRequest(http.MethodGet, "/api/items?category=Bug+Bounty+%26+Security", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var catFiltered []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&catFiltered); err != nil {
		t.Fatal(err)
	}
	if len(catFiltered) != 1 || catFiltered[0].Name != "security-audit" {
		t.Fatalf("expected security-audit for Bug Bounty category, got %v", catFiltered)
	}
}

func TestHandleGetItemsWithFamily(t *testing.T) {
	srv, svc := setupTestServer()
	svc.SetItems([]domain.Item{
		{ID: "1", Name: "vue", Type: domain.ItemTypeSkill},
		{ID: "2", Name: "vue-best-practices", Type: domain.ItemTypeSkill},
		{ID: "3", Name: "hunt-xss", Type: domain.ItemTypeSkill},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/items?family=vue", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var items []domain.Item
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 vue items, got %d", len(items))
	}
	for _, it := range items {
		if it.Family != "vue" {
			t.Errorf("expected family vue, got %s", it.Family)
		}
	}
}

func TestHandleGetStats(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var stats domain.Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if stats.Total != 2 {
		t.Errorf("expected total 2, got %d", stats.Total)
	}
	if stats.ByType[domain.ItemTypeMCP] != 1 {
		t.Errorf("expected 1 MCP, got %d", stats.ByType[domain.ItemTypeMCP])
	}
	if stats.ClassifiedCount != 1 {
		t.Errorf("expected 1 classified, got %d", stats.ClassifiedCount)
	}
	if stats.UnclassifiedCount != 1 {
		t.Errorf("expected 1 unclassified, got %d", stats.UnclassifiedCount)
	}
}

func TestHandleTaxonomyEndpoints(t *testing.T) {
	srv, svc := setupTestServer()
	svc.GetTaxonomy().SetCustomFilePath(filepath.Join(t.TempDir(), "taxonomy.json"))

	// GET /api/taxonomy
	req := httptest.NewRequest(http.MethodGet, "/api/taxonomy", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var taxResp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&taxResp); err != nil {
		t.Fatal(err)
	}
	categories, ok := taxResp["categories"].([]any)
	if !ok || len(categories) == 0 {
		t.Fatalf("expected categories array, got %v", taxResp["categories"])
	}

	// POST /api/taxonomy
	body := map[string]any{
		"category":    "Code Analysis",
		"subCategory": "AST & Graphs",
		"patterns":    []string{"unknown-mcp*"},
	}
	bodyBytes, _ := json.Marshal(body)
	postReq := httptest.NewRequest(http.MethodPost, "/api/taxonomy", bytes.NewReader(bodyBytes))
	postRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body: %s)", postRec.Code, postRec.Body.String())
	}

	// Verify that unknown-mcp item is now classified under "Code Analysis"
	itemsReq := httptest.NewRequest(http.MethodGet, "/api/items?category=Code+Analysis", nil)
	itemsRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(itemsRec, itemsReq)

	var items []domain.Item
	if err := json.NewDecoder(itemsRec.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "unknown-mcp" {
		t.Fatalf("expected unknown-mcp to be reclassified under Code Analysis, got %v", items)
	}
}

func TestCORSHeaders(t *testing.T) {
	srv, _ := setupTestServer()

	// Preflight OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/api/items", nil)
	req.Header.Set("Origin", "http://localhost:4321")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 for OPTIONS, got %d", rec.Code)
	}

	originHeader := rec.Header().Get("Access-Control-Allow-Origin")
	if originHeader != "http://localhost:4321" {
		t.Fatalf("expected allow origin http://localhost:4321, got %s", originHeader)
	}
}
