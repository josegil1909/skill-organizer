package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestGrokScanner(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	content := `
[mcp_servers.stitch]
command = "npx"
args = ["-y", "@_davideast/stitch-mcp"]
enabled = true

[mcp_servers.stitch.env]
STITCH_KEY = "super-secret"
`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewGrokScanner(configFile)
	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	it := items[0]
	if it.Name != "stitch" {
		t.Errorf("expected name 'stitch', got '%s'", it.Name)
	}
	if it.Provider != domain.ProviderGrok {
		t.Errorf("expected provider grok, got '%s'", it.Provider)
	}
	if it.Invocation != "npx -y @_davideast/stitch-mcp" {
		t.Errorf("unexpected invocation: %s", it.Invocation)
	}
	if len(it.EnvKeys) != 1 || it.EnvKeys[0] != "STITCH_KEY" {
		t.Errorf("expected envKeys [STITCH_KEY], got %v", it.EnvKeys)
	}
}
