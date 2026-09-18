package mcp

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"organizer/backend/internal/adapter"
	"organizer/backend/internal/domain"
)

// HermesScanner scans Hermes configuration (~/.hermes/config.yaml).
type HermesScanner struct {
	configPath string
}

// NewHermesScanner creates a new Hermes scanner.
func NewHermesScanner(configPath ...string) *HermesScanner {
	p := "~/.hermes/config.yaml"
	if len(configPath) > 0 && configPath[0] != "" {
		p = configPath[0]
	}
	return &HermesScanner{configPath: p}
}

// Name returns the identifier of this adapter.
func (s *HermesScanner) Name() string {
	return "hermes-mcp-scanner"
}

type hermesConfigFile struct {
	MCPServers map[string]hermesMCPEntry `yaml:"mcp_servers"`
}

type hermesMCPEntry struct {
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Environment map[string]string `yaml:"environment"`
	URL         string            `yaml:"url"`
	Transport   string            `yaml:"transport"`
	Timeout     int               `yaml:"timeout"`
	Enabled     *bool             `yaml:"enabled"`
}

// Scan scans ~/.hermes/config.yaml and extracts domain Items.
func (s *HermesScanner) Scan(ctx context.Context) ([]domain.Item, error) {
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

	var cfg hermesConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, nil
	}

	var items []domain.Item
	for name, entry := range cfg.MCPServers {
		inv := adapter.FormatInvocation(entry.Command, entry.Args, entry.URL, entry.Transport)

		envMap := entry.Env
		if len(envMap) == 0 && len(entry.Environment) > 0 {
			envMap = entry.Environment
		}

		desc := fmt.Sprintf("Hermes MCP server '%s'", name)
		if entry.URL != "" {
			desc = fmt.Sprintf("Hermes remote MCP server at %s", entry.URL)
		} else if entry.Command != "" {
			desc = fmt.Sprintf("Hermes local MCP server via %s", entry.Command)
		}
		if entry.Enabled != nil && !*entry.Enabled {
			desc += " (disabled)"
		}

		id := fmt.Sprintf("%s:%s:%s", domain.ProviderHermes, domain.ItemTypeMCP, name)

		item := domain.Item{
			ID:          id,
			Name:        name,
			Type:        domain.ItemTypeMCP,
			Provider:    domain.ProviderHermes,
			Origin:      adapter.ClassifyMCPOrigin(entry.Command, entry.Args, entry.URL),
			SourcePath:  filePath,
			Description: desc,
			Command:     entry.Command,
			Args:        entry.Args,
			EnvKeys:     adapter.ExtractEnvKeys(envMap),
			URL:         entry.URL,
			Invocation:  inv,
			RawConfig:   adapter.SanitizeConfigJSON(entry),
		}

		items = append(items, item)
	}

	return items, nil
}
