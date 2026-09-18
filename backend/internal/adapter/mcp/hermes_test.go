package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestHermesScanner(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	content := `
mcp_servers:
  codegraph:
    command: codegraph
    args:
    - serve
    - --mcp
    timeout: 120
    connect_timeout: 60
    enabled: true
    env:
      TOKEN: secret-xyz
`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewHermesScanner(configFile)
	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	it := items[0]
	if it.Name != "codegraph" {
		t.Errorf("expected name 'codegraph', got '%s'", it.Name)
	}
	if it.Provider != domain.ProviderHermes {
		t.Errorf("expected provider hermes, got '%s'", it.Provider)
	}
	if it.Invocation != "codegraph serve --mcp" {
		t.Errorf("unexpected invocation: %s", it.Invocation)
	}
	if len(it.EnvKeys) != 1 || it.EnvKeys[0] != "TOKEN" {
		t.Errorf("expected envKeys [TOKEN], got %v", it.EnvKeys)
	}
}
