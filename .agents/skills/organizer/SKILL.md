---
name: organizer-catalog
description: Explore, inspect, and manage skills and MCP servers across all local AI clients (Grok, Hermes, OpenCode, Cursor, etc.). Helps classify unclassified tools and discover how to invoke them.
---

# Organizer AI Catalog & MCP Server

Organizer is the unified local control center and registry for AI skills and Model Context Protocol (MCP) servers across all installed AI ecosystems:
- **Cursor** (`~/.cursor/mcp.json`, `~/.cursor/skills/`)
- **Claude Desktop** (`~/.config/Claude/claude_desktop_config.json`)
- **OpenCode** (`~/.config/opencode/mcp.json`)
- **Hermes** (`~/.hermes/mcp.json`)
- **Grok** (`~/.grok/config.toml`, `~/.grok/mcp.json`)
- **Kimi** (`~/.kimi/mcp.json`)
- **Pi** (`~/.pi/agent/skills/`)
- **Global Skills** (`~/.agents/skills/`)

---

## 1. Using the Organizer Built-in MCP Server (stdio)

You can connect directly to Organizer as a standard Model Context Protocol (MCP) stdio server.

### Launch Command
```bash
# From a clone of this repo
go run ./backend mcp

# After `make build`
./bin/organizer-backend mcp
```

### Configuration Snippet (for Cursor / Claude / Antigravity)
```json
{
  "mcpServers": {
    "organizer": {
      "command": "/absolute/path/to/organizer/bin/organizer-backend",
      "args": ["mcp"]
    }
  }
}
```

### Available MCP Tools
1. **`organizer_list_items`**:
   - Query catalog items with filtering:
     - `type`: `"skill"` or `"mcp"`
     - `provider`: `"grok"`, `"hermes"`, `"cursor"`, `"claude"`, `"opencode"`, `"global"`, etc.
     - `category`: e.g. `"Bug Bounty & Security"`, `"Gentle AI & SDD"`, `"Frontend & UI/UX"`
     - `subCategory`: e.g. `"Web Vulnerabilities"`, `"Spec-Driven Development"`
     - `isClassified`: `true` or `false`
     - `search`: search substring across name, description, command, invocation
     - `limit`: max results (default 50)
2. **`organizer_get_item`**:
   - Inspect item by ID (e.g. `id: "grok:mcp:obsidian"`, `id: "global:skill:hunt-sqli"`).
   - Returns invocation syntax, environment variable requirements, execution command, and raw configuration.
3. **`organizer_get_unclassified`**:
   - Lists items currently pending classification triage (`isClassified: false`, `Category: "Unclassified"`).
   - Used by AI assistants to discover items needing new taxonomy rules.
4. **`organizer_add_taxonomy_rule`**:
   - Add a declarative rule dynamically to categorize items:
     - `category`: target suite (e.g. `"Bug Bounty & Security"`)
     - `subCategory`: subcategory (e.g. `"Web Vulnerabilities"`)
     - `pattern` or `patterns`: glob patterns (e.g. `["sqli-*", "xss-*"]`)
   - Immediately reclassifies all items in the catalog and persists to `~/.config/organizer/taxonomy.json`.
5. **`organizer_get_stats`**:
   - Aggregated catalog statistics: total items, classified vs unclassified counts, distribution by provider, type, origin, and categories.
6. **`organizer_refresh`**:
   - Re-scans all AI client configuration directories on disk and returns updated counts.

---

## 2. HTTP REST API Reference

When the backend is running in HTTP mode (`go run main.go` or `organizer-backend` on port `8080`):

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/health` | Healthcheck (`{"status": "ok"}`) |
| `GET` | `/api/items` | Query items. Parameters: `?type=...&provider=...&origin=...&category=...&subCategory=...&isClassified=true/false&search=...` |
| `GET` | `/api/stats` | Aggregated statistics |
| `POST` | `/api/refresh` | Trigger re-scan of client configs |
| `GET` | `/api/taxonomy` | Get active declarative rules and categories |
| `POST` | `/api/taxonomy` | Add a classification rule (`{"category": "...", "subCategory": "...", "patterns": [...]}`) |

---

## 3. How to Triage Unclassified Items

When new tools or custom skills are added to the system, they start with:
- `category`: `"Unclassified"`
- `subCategory`: `"Pending Triage"`
- `isClassified`: `false`

### Step-by-Step Triage Workflow for AI:
1. Call `organizer_get_unclassified(limit=20)`.
2. Inspect the unclassified items' names, source paths, and descriptions.
3. Determine appropriate `category` and `subCategory` according to the taxonomy suites:
   - **Bug Bounty & Security**: Web Vulnerabilities, Recon & OSINT, Red Team & Exploits, Audit & Compliance
   - **Gentle AI & SDD**: Spec-Driven Development, Review & Benchmark
   - **Frontend & UI/UX**: Aesthetics & Taste, Design Systems & Tokens, Redesign & Polish, Visual AI & Image-to-Code, Web Standards & Testing
   - **Workflow & Engineering**: Git & PR Automation, Code Quality & Testing
   - **AI Core & Meta**: Skill Management
   - **Cloud & Integrations**: Productivity & Services, Developer Tools & Infra
4. Call `organizer_add_taxonomy_rule` with suitable glob pattern(s).
5. The catalog is instantly re-classified and saved.
