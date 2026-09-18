package adapter

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtractEnvKeys(t *testing.T) {
	env := map[string]string{
		"SECRET_TOKEN": "12345",
		"API_KEY":      "abcd",
		"PORT":         "8080",
	}

	keys := ExtractEnvKeys(env)
	expected := []string{"API_KEY", "PORT", "SECRET_TOKEN"}
	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("expected %v, got %v", expected, keys)
	}

	if ExtractEnvKeys(nil) != nil {
		t.Fatal("expected nil for empty map")
	}
}

func TestFormatInvocation(t *testing.T) {
	inv := FormatInvocation("npx", []string{"-y", "@mcp/server"}, "", "")
	if inv != "npx -y @mcp/server" {
		t.Fatalf("unexpected invocation: %s", inv)
	}

	invURL := FormatInvocation("", nil, "https://mcp.example.com", "http")
	if invURL != "https://mcp.example.com (http)" {
		t.Fatalf("unexpected url invocation: %s", invURL)
	}
}

func TestSanitizeConfigJSON(t *testing.T) {
	config := map[string]any{
		"command": "serve",
		"env": map[string]string{
			"SECRET": "super-secret-value",
		},
	}

	out := SanitizeConfigJSON(config)
	if strings.Contains(out, "super-secret-value") {
		t.Fatal("secret value leaked in sanitized JSON output")
	}
	if !strings.Contains(out, "***") {
		t.Fatal("expected masked env value '***'")
	}
}

func TestClassifyMCPOrigin(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []string
		url      string
		expected string
	}{
		{"npx runner", "npx", []string{"-y", "pkg"}, "", "installed"},
		{"uvx runner", "uvx", []string{"pkg"}, "", "installed"},
		{"docker runner", "docker", []string{"run", "image"}, "", "installed"},
		{"system path", "/usr/bin/tool", nil, "", "installed"},
		{"homebrew path", "/home/linuxbrew/.linuxbrew/bin/engram", nil, "", "installed"},
		{"global binary", "codegraph", nil, "", "installed"},
		{"local script relative", "./script.sh", nil, "", "custom"},
		{"python user script", "python3", []string{"/home/user/dev/server.py"}, "", "custom"},
		{"remote url", "", nil, "https://mcp.context7.com/mcp", "installed"},
		{"localhost url", "", nil, "http://localhost:3000/mcp", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyMCPOrigin(tt.command, tt.args, tt.url)
			if got != tt.expected {
				t.Errorf("ClassifyMCPOrigin(%q, %v, %q) = %q, expected %q", tt.command, tt.args, tt.url, got, tt.expected)
			}
		})
	}
}

