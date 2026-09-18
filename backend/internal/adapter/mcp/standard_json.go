package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
)

// StandardJSONTarget defines a JSON MCP config file and its provider.
type StandardJSONTarget struct {
	Path     string
	Provider domain.Provider
}

// DefaultStandardJSONTargets contains standard JSON MCP config paths.
var DefaultStandardJSONTargets = []StandardJSONTarget{
	{Path: "~/.cursor/mcp.json", Provider: domain.ProviderCursor},
	{Path: "~/.claude.json", Provider: domain.ProviderClaude},
	{Path: "~/.kimi-code/mcp.json", Provider: domain.ProviderKimi},
}

// StandardJSONScanner scans standard JSON configs containing an "mcpServers" map.
type StandardJSONScanner struct {
	targets []StandardJSONTarget
}

// NewStandardJSONScanner creates a new standard JSON MCP scanner.
func NewStandardJSONScanner(targets ...StandardJSONTarget) *StandardJSONScanner {
	if len(targets) == 0 {
		targets = DefaultStandardJSONTargets
	}
	return &StandardJSONScanner{targets: targets}
}

// Name returns the identifier of this adapter.
func (s *StandardJSONScanner) Name() string {
	return "standard-json-mcp-scanner"
}

type standardConfigFile struct {
	MCPServers map[string]standardMCPEntry `json:"mcpServers"`
}

type standardMCPEntry struct {
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	URL       string            `json:"url"`
	Transport string            `json:"transport"`
	Type      string            `json:"type"`
	Disabled  *bool             `json:"disabled"`
}

// Scan scans all configured JSON targets and extracts domain Items.
func (s *StandardJSONScanner) Scan(ctx context.Context) ([]domain.Item, error) {
	var allItems []domain.Item

	for _, target := range s.targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		filePath := adapter.ExpandHome(target.Path)
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Skip gracefully if file does not exist or cannot be read
			continue
		}

		var cfg standardConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		for name, entry := range cfg.MCPServers {
			transport := entry.Transport
			if transport == "" {
				transport = entry.Type
			}

			inv := adapter.FormatInvocation(entry.Command, entry.Args, entry.URL, transport)

			desc := fmt.Sprintf("MCP server '%s'", name)
			if entry.URL != "" {
				desc = fmt.Sprintf("Remote MCP server at %s", entry.URL)
			} else if entry.Command != "" {
				desc = fmt.Sprintf("Local MCP server via %s", entry.Command)
			}

			id := fmt.Sprintf("%s:%s:%s", target.Provider, domain.ItemTypeMCP, name)

			item := domain.Item{
				ID:          id,
				Name:        name,
				Type:        domain.ItemTypeMCP,
				Provider:    target.Provider,
				Origin:      adapter.ClassifyMCPOrigin(entry.Command, entry.Args, entry.URL),
				SourcePath:  filePath,
				Description: desc,
				Command:     entry.Command,
				Args:        entry.Args,
				EnvKeys:     adapter.ExtractEnvKeys(entry.Env),
				URL:         entry.URL,
				Invocation:  inv,
				RawConfig:   adapter.SanitizeConfigJSON(entry),
			}

			allItems = append(allItems, item)
		}
	}

	return allItems, nil
}
