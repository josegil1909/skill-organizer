package domain

// ItemType defines whether an item is a skill or an MCP server.
type ItemType string

const (
	ItemTypeSkill ItemType = "skill"
	ItemTypeMCP   ItemType = "mcp"
)

// Provider represents the source AI client or ecosystem.
type Provider string

const (
	ProviderGlobal   Provider = "global"
	ProviderCursor   Provider = "cursor"
	ProviderClaude   Provider = "claude"
	ProviderKimi     Provider = "kimi"
	ProviderOpenCode Provider = "opencode"
	ProviderHermes   Provider = "hermes"
	ProviderGrok     Provider = "grok"
	ProviderGemini   Provider = "gemini"
	ProviderPi       Provider = "pi"
)

// Origin defines whether an item is builtin, installed, or custom.
const (
	OriginBuiltin   = "builtin"
	OriginInstalled = "installed"
	OriginCustom    = "custom"
)

// Item represents a scanned skill or MCP server.
type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        ItemType `json:"type"`
	Provider    Provider `json:"provider"`
	Origin      string   `json:"origin"`
	SourcePath  string   `json:"sourcePath"`
	Description string   `json:"description"`
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	EnvKeys     []string `json:"envKeys,omitempty"` // DO NOT expose secret values, only keys!
	URL         string   `json:"url,omitempty"`
	Category    string   `json:"category"`
	SubCategory string   `json:"subCategory"`
	IsClassified bool    `json:"isClassified"`
	SourceURL   string   `json:"sourceUrl,omitempty"`
	SourcePkg   string   `json:"sourcePkg,omitempty"`   // Source package/repo (e.g. "vercel-labs/skills")
	RegistryURL string   `json:"registryUrl,omitempty"` // Public registry URL on skills.sh
	Invocation  string   `json:"invocation"` // How to call/use it
	RawConfig   string   `json:"rawConfig,omitempty"`
	Family      string   `json:"family,omitempty"`     // Family / Parent slug (e.g., "vue", "hunt", "sdd", "game", "golang")
	IsParent    bool     `json:"isParent,omitempty"`   // True if this skill is the parent / root of a family
	ChildCount  int      `json:"childCount,omitempty"` // Number of sub-skills in this family
}

// Stats provides aggregated summary counts of scanned items.
type Stats struct {
	Total             int              `json:"total"`
	ByType            map[ItemType]int `json:"byType"`
	ByProvider        map[Provider]int `json:"byProvider"`
	ByOrigin          map[string]int   `json:"byOrigin"`
	ByCategory        map[string]int   `json:"byCategory"`
	BySubCategory     map[string]int   `json:"bySubCategory"`
	ByFamily          map[string]int   `json:"byFamily,omitempty"`
	ClassifiedCount   int              `json:"classifiedCount"`
	UnclassifiedCount int              `json:"unclassifiedCount"`
}
