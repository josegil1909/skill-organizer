package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestStandardJSONScanner(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "mcp.json")

	content := `{
  "mcpServers": {
    "test-cmd": {
      "command": "npx",
      "args": ["-y", "test-server"],
      "env": {
        "SECRET_KEY": "supersecret"
      }
    },
    "test-remote": {
      "transport": "http",
      "url": "https://mcp.example.com/api"
    }
  }
}`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewStandardJSONScanner(StandardJSONTarget{
		Path:     configFile,
		Provider: domain.ProviderCursor,
	})

	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	var foundCmd, foundRemote bool
	for _, it := range items {
		if it.Name == "test-cmd" {
			foundCmd = true
			if it.Command != "npx" {
				t.Errorf("expected command npx, got %s", it.Command)
			}
			if len(it.EnvKeys) != 1 || it.EnvKeys[0] != "SECRET_KEY" {
				t.Errorf("expected envKeys [SECRET_KEY], got %v", it.EnvKeys)
			}
			if it.Invocation != "npx -y test-server" {
				t.Errorf("unexpected invocation: %s", it.Invocation)
			}
		}
		if it.Name == "test-remote" {
			foundRemote = true
			if it.URL != "https://mcp.example.com/api" {
				t.Errorf("expected url, got %s", it.URL)
			}
			if it.Invocation != "https://mcp.example.com/api (http)" {
				t.Errorf("unexpected invocation: %s", it.Invocation)
			}
		}
	}

	if !foundCmd || !foundRemote {
		t.Fatalf("expected both items to be found")
	}
}
