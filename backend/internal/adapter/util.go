package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"organizer/backend/internal/domain"
)

// ExpandHome replaces a leading "~" or "~/" with the user's home directory.
func ExpandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// ExtractEnvKeys returns sorted keys from an environment variable map, guaranteeing no secret values are exposed.
func ExtractEnvKeys(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FormatInvocation constructs a human-readable invocation hint for a command or remote URL.
func FormatInvocation(command string, args []string, url string, transport string) string {
	if command != "" {
		if len(args) > 0 {
			return fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		}
		return command
	}
	if url != "" {
		if transport != "" && transport != "http" && transport != "sse" {
			return fmt.Sprintf("%s (%s)", url, transport)
		}
		if transport != "" {
			return fmt.Sprintf("%s (%s)", url, transport)
		}
		return url
	}
	return ""
}

// FormatSkillInvocation formats an invocation hint for a skill.
func FormatSkillInvocation(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("/%s or @%s", name, name)
}

// SanitizeConfigJSON serializes any configuration data to indented JSON while masking any secret env values.
func SanitizeConfigJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return string(data)
	}

	maskEnvInMap(raw)

	sanitized, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(sanitized)
}

func maskEnvInMap(m map[string]any) {
	for k, val := range m {
		lowerK := strings.ToLower(k)
		if lowerK == "env" || lowerK == "environment" {
			if envMap, ok := val.(map[string]any); ok {
				masked := make(map[string]any, len(envMap))
				for envKey := range envMap {
					masked[envKey] = "***"
				}
				m[k] = masked
				continue
			}
		}
		if subMap, ok := val.(map[string]any); ok {
			maskEnvInMap(subMap)
		}
	}
}

// ClassifyMCPOrigin classifies an MCP server origin as "installed" or "custom".
func ClassifyMCPOrigin(command string, args []string, url string) string {
	if url != "" {
		if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
			return domain.OriginCustom
		}
		return domain.OriginInstalled
	}

	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return domain.OriginCustom
	}

	cmdLower := strings.ToLower(cmd)
	baseName := strings.ToLower(filepath.Base(cmd))

	// Package / container runners indicate an installed package or tool
	if baseName == "npx" || baseName == "uvx" || baseName == "pipx" || baseName == "bunx" || baseName == "docker" {
		return domain.OriginInstalled
	}

	// Script interpreter running a user script file
	if baseName == "python" || baseName == "python3" || baseName == "node" || baseName == "deno" || baseName == "bun" || baseName == "bash" || baseName == "sh" {
		for _, arg := range args {
			argLower := strings.ToLower(arg)
			if strings.HasSuffix(argLower, ".py") || strings.HasSuffix(argLower, ".js") || strings.HasSuffix(argLower, ".ts") || strings.HasSuffix(argLower, ".sh") {
				if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "../") || strings.HasPrefix(arg, "/") {
					return domain.OriginCustom
				}
			}
		}
	}

	// System or package manager paths
	if strings.HasPrefix(cmdLower, "/usr/") ||
		strings.HasPrefix(cmdLower, "/home/linuxbrew/") ||
		strings.HasPrefix(cmdLower, "/opt/") ||
		strings.Contains(cmdLower, "/.local/bin/") ||
		strings.Contains(cmdLower, "/node_modules/.bin/") {
		return domain.OriginInstalled
	}

	// Local relative paths
	if strings.HasPrefix(cmd, "./") || strings.HasPrefix(cmd, "../") {
		return domain.OriginCustom
	}

	// Script extensions (e.g. .py, .sh, .js, .ts)
	if strings.HasSuffix(baseName, ".py") ||
		strings.HasSuffix(baseName, ".sh") ||
		strings.HasSuffix(baseName, ".js") ||
		strings.HasSuffix(baseName, ".ts") {
		return domain.OriginCustom
	}

	// Global binary names in PATH (e.g. "codegraph", "git", "sqlite3")
	if !strings.Contains(cmd, "/") && !strings.Contains(cmd, "\\") {
		return domain.OriginInstalled
	}

	return domain.OriginCustom
}

