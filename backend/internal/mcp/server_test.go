package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"organizer/backend/internal/domain"
	"organizer/backend/internal/service"
	"organizer/backend/internal/taxonomy"
)

func setupTestMCPServer() (*Server, *service.AggregatorService, *bytes.Buffer) {
	svc := service.NewAggregatorService()
	svc.SetTaxonomy(taxonomy.NewDefaultEngine())
	svc.SetItems([]domain.Item{
		{
			ID:          "cursor:mcp:filesystem",
			Name:        "filesystem",
			Type:        domain.ItemTypeMCP,
			Provider:    domain.ProviderCursor,
			Origin:      domain.OriginInstalled,
			Description: "Local filesystem server",
			Category:    "Cloud & Integrations",
			SubCategory: "Developer Tools & Infra",
			IsClassified: true,
		},
		{
			ID:          "global:skill:unknown-tool",
			Name:        "unknown-tool",
			Type:        domain.ItemTypeSkill,
			Provider:    domain.ProviderGlobal,
			Origin:      domain.OriginCustom,
			Description: "Random unclassified script",
			Category:    "Unclassified",
			SubCategory: "Pending Triage",
			IsClassified: false,
		},
	})

	outBuf := new(bytes.Buffer)
	srv := NewServer(svc, nil, outBuf)
	return srv, svc, outBuf
}

func TestMCPInitialize(t *testing.T) {
	srv, _, outBuf := setupTestMCPServer()

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}

	srv.handleRequest(context.Background(), req)

	var resp RPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if string(resp.ID) != "1" {
		t.Errorf("expected ID 1, got %s", string(resp.ID))
	}
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if resultMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocolVersion: %v", resultMap["protocolVersion"])
	}
}

func TestMCPToolsList(t *testing.T) {
	srv, _, outBuf := setupTestMCPServer()

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}

	srv.handleRequest(context.Background(), req)

	var resp RPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	tools, ok := resultMap["tools"].([]any)
	if !ok || len(tools) != 6 {
		t.Errorf("expected 6 tools, got %v", resultMap["tools"])
	}
}

func extractToolContentText(t *testing.T, respBytes []byte) string {
	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Result.Content) == 0 {
		t.Fatalf("expected at least 1 content block, got 0")
	}
	return resp.Result.Content[0].Text
}

func TestMCPToolCallGetStats(t *testing.T) {
	srv, _, outBuf := setupTestMCPServer()

	params := map[string]any{
		"name":      "organizer_get_stats",
		"arguments": map[string]any{},
	}
	paramsBytes, _ := json.Marshal(params)

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  paramsBytes,
	}

	srv.handleRequest(context.Background(), req)

	text := extractToolContentText(t, outBuf.Bytes())
	if !strings.Contains(text, `"total": 2`) {
		t.Errorf("expected total 2 in stats text: %s", text)
	}
}

func TestMCPToolCallGetUnclassified(t *testing.T) {
	srv, _, outBuf := setupTestMCPServer()

	params := map[string]any{
		"name":      "organizer_get_unclassified",
		"arguments": map[string]any{"limit": 10},
	}
	paramsBytes, _ := json.Marshal(params)

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "tools/call",
		Params:  paramsBytes,
	}

	srv.handleRequest(context.Background(), req)

	text := extractToolContentText(t, outBuf.Bytes())
	if !strings.Contains(text, "unknown-tool") {
		t.Errorf("expected unknown-tool in unclassified response: %s", text)
	}
}

func TestMCPToolCallAddTaxonomyRule(t *testing.T) {
	srv, svc, outBuf := setupTestMCPServer()
	svc.GetTaxonomy().SetCustomFilePath(filepath.Join(t.TempDir(), "taxonomy.json"))

	params := map[string]any{
		"name": "organizer_add_taxonomy_rule",
		"arguments": map[string]any{
			"category":    "Experimental",
			"subCategory": "Unknown Scripts",
			"patterns":    []string{"unknown-*"},
		},
	}
	paramsBytes, _ := json.Marshal(params)

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`5`),
		Method:  "tools/call",
		Params:  paramsBytes,
	}

	srv.handleRequest(context.Background(), req)

	text := extractToolContentText(t, outBuf.Bytes())
	if !strings.Contains(text, `"status": "ok"`) {
		t.Errorf("expected status ok: %s", text)
	}
}

func TestMCPServeStream(t *testing.T) {
	svc := service.NewAggregatorService()
	inBuf := bytes.NewBufferString(`{"jsonrpc":"2.0","id":10,"method":"ping"}` + "\n")
	outBuf := new(bytes.Buffer)

	srv := NewServer(svc, inBuf, outBuf)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := srv.Serve(ctx)
	if err != nil {
		t.Fatalf("unexpected Serve error: %v", err)
	}

	if !strings.Contains(outBuf.String(), `"id":10`) {
		t.Errorf("expected response to ping with id 10, got %s", outBuf.String())
	}
}
