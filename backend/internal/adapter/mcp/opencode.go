package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
)

// OpenCodeScanner scans OpenCode configuration (~/.config/opencode/opencode.json).
type OpenCodeScanner struct {
	configPath string
}

// NewOpenCodeScanner creates a new OpenCode scanner.
func NewOpenCodeScanner(configPath ...string) *OpenCodeScanner {
	p := "~/.config/opencode/opencode.json"
	if len(configPath) > 0 && configPath[0] != "" {
		p = configPath[0]
	}
	return &OpenCodeScanner{configPath: p}
}

// Name returns the identifier of this adapter.
func (s *OpenCodeScanner) Name() string {
	return "opencode-mcp-scanner"
}

type openCodeConfigFile struct {
	MCP map[string]openCodeMCPEntry `json:"mcp"`
}

type openCodeMCPEntry struct {
	RawCommand  any               `json:"command"`
	Args        []string          `json:"args"`
	Type        string            `json:"type"`
	Enabled     *bool             `json:"enabled"`
	Environment map[string]string `json:"environment"`
	URL         string            `json:"url"`
}

// Scan scans the OpenCode configuration and extracts domain Items.
func (s *OpenCodeScanner) Scan(ctx context.Context) ([]domain.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath := adapter.ExpandHome(s.configPath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Gracefully skip if config file is missing
		return nil, nil
	}

	var cfg openCodeConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil
	}

	var items []domain.Item
	for name, entry := range cfg.MCP {
		cmd, args := parseOpenCodeCommand(entry.RawCommand, entry.Args)

		inv := adapter.FormatInvocation(cmd, args, entry.URL, entry.Type)

		desc := fmt.Sprintf("OpenCode MCP server '%s'", name)
		if entry.URL != "" {
			desc = fmt.Sprintf("OpenCode remote MCP server at %s", entry.URL)
		} else if cmd != "" {
			desc = fmt.Sprintf("OpenCode local MCP server via %s", cmd)
		}
		if entry.Enabled != nil && !*entry.Enabled {
			desc += " (disabled)"
		}

		id := fmt.Sprintf("%s:%s:%s", domain.ProviderOpenCode, domain.ItemTypeMCP, name)

		item := domain.Item{
			ID:          id,
			Name:        name,
			Type:        domain.ItemTypeMCP,
			Provider:    domain.ProviderOpenCode,
			Origin:      adapter.ClassifyMCPOrigin(cmd, args, entry.URL),
			SourcePath:  filePath,
			Description: desc,
			Command:     cmd,
			Args:        args,
			EnvKeys:     adapter.ExtractEnvKeys(entry.Environment),
			URL:         entry.URL,
			Invocation:  inv,
			RawConfig:   adapter.SanitizeConfigJSON(entry),
		}

		items = append(items, item)
	}

	return items, nil
}

func parseOpenCodeCommand(raw any, extraArgs []string) (string, []string) {
	var cmd string
	var args []string

	switch v := raw.(type) {
	case string:
		cmd = v
		args = append(args, extraArgs...)
	case []any:
		for i, item := range v {
			if s, ok := item.(string); ok {
				if i == 0 {
					cmd = s
				} else {
					args = append(args, s)
				}
			}
		}
		args = append(args, extraArgs...)
	case []string:
		if len(v) > 0 {
			cmd = v[0]
			args = append(args, v[1:]...)
		}
		args = append(args, extraArgs...)
	}

	return cmd, args
}
