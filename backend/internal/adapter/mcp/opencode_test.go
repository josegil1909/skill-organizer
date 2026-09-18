package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"organizer/backend/internal/domain"
)

func TestOpenCodeScanner(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "opencode.json")

	content := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "local-server": {
      "command": ["codegraph", "serve", "--mcp"],
      "type": "local",
      "enabled": true
    },
    "remote-server": {
      "type": "remote",
      "url": "https://mcp.context7.com/mcp",
      "enabled": true
    },
    "custom-env": {
      "command": "/bin/tool",
      "args": ["--run"],
      "environment": {
        "API_SECRET": "12345"
      },
      "type": "local"
    }
  }
}`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewOpenCodeScanner(configFile)
	items, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	var foundLocal, foundRemote, foundCustom bool
	for _, it := range items {
		if it.Provider != domain.ProviderOpenCode {
			t.Errorf("expected provider opencode, got %s", it.Provider)
		}
		if it.Name == "local-server" {
			foundLocal = true
			if it.Command != "codegraph" {
				t.Errorf("expected command 'codegraph', got '%s'", it.Command)
			}
			if len(it.Args) != 2 || it.Args[0] != "serve" || it.Args[1] != "--mcp" {
				t.Errorf("expected args ['serve', '--mcp'], got %v", it.Args)
			}
			if it.Invocation != "codegraph serve --mcp" {
				t.Errorf("unexpected invocation: %s", it.Invocation)
			}
		}
		if it.Name == "remote-server" {
			foundRemote = true
			if it.URL != "https://mcp.context7.com/mcp" {
				t.Errorf("expected url, got %s", it.URL)
			}
			if it.Invocation != "https://mcp.context7.com/mcp (remote)" {
				t.Errorf("unexpected invocation: %s", it.Invocation)
			}
		}
		if it.Name == "custom-env" {
			foundCustom = true
			if len(it.EnvKeys) != 1 || it.EnvKeys[0] != "API_SECRET" {
				t.Errorf("expected envKeys [API_SECRET], got %v", it.EnvKeys)
			}
			if it.Invocation != "/bin/tool --run" {
				t.Errorf("unexpected invocation: %s", it.Invocation)
			}
		}
	}

	if !foundLocal || !foundRemote || !foundCustom {
		t.Fatalf("not all items were found")
	}
}
