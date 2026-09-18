package mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
)

// GrokScanner scans Grok configuration (~/.grok/config.toml).
type GrokScanner struct {
	configPath string
}

// NewGrokScanner creates a new Grok scanner.
func NewGrokScanner(configPath ...string) *GrokScanner {
	p := "~/.grok/config.toml"
	if len(configPath) > 0 && configPath[0] != "" {
		p = configPath[0]
	}
	return &GrokScanner{configPath: p}
}

// Name returns the identifier of this adapter.
func (s *GrokScanner) Name() string {
	return "grok-mcp-scanner"
}

type grokConfigFile struct {
	MCPServers map[string]grokMCPEntry `toml:"mcp_servers"`
}

type grokMCPEntry struct {
	Command           string            `toml:"command"`
	Args              []string          `toml:"args"`
	Env               map[string]string `toml:"env"`
	URL               string            `toml:"url"`
	Transport         string            `toml:"transport"`
	Enabled           *bool             `toml:"enabled"`
	StartupTimeoutSec int               `toml:"startup_timeout_sec"`
}

// Scan scans ~/.grok/config.toml and extracts domain Items.
func (s *GrokScanner) Scan(ctx context.Context) ([]domain.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath := adapter.ExpandHome(s.configPath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Gracefully skip missing config file
		return nil, nil
	}

	var cfg grokConfigFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, nil
	}

	var items []domain.Item
	for name, entry := range cfg.MCPServers {
		inv := adapter.FormatInvocation(entry.Command, entry.Args, entry.URL, entry.Transport)

		desc := fmt.Sprintf("Grok MCP server '%s'", name)
		if entry.URL != "" {
			desc = fmt.Sprintf("Grok remote MCP server at %s", entry.URL)
		} else if entry.Command != "" {
			desc = fmt.Sprintf("Grok local MCP server via %s", entry.Command)
		}
		if entry.Enabled != nil && !*entry.Enabled {
			desc += " (disabled)"
		}

		id := fmt.Sprintf("%s:%s:%s", domain.ProviderGrok, domain.ItemTypeMCP, name)

		item := domain.Item{
			ID:          id,
			Name:        name,
			Type:        domain.ItemTypeMCP,
			Provider:    domain.ProviderGrok,
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

		items = append(items, item)
	}

	return items, nil
}
