package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"organizer/backend/internal/domain"
	"organizer/backend/internal/service"
)

// Server implements a stdio JSON-RPC 2.0 MCP server.
type Server struct {
	svc      *service.AggregatorService
	in       io.Reader
	out      io.Writer
	outMu    sync.Mutex
	errLog   *log.Logger
	tools    []Tool
}

// Tool defines an MCP tool definition.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// RPCRequest represents an incoming JSON-RPC 2.0 message.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCResponse represents an outgoing JSON-RPC 2.0 response.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError defines a standard JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// TextContent represents text block in MCP tool response.
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolCallResult represents the payload returned by tools/call.
type ToolCallResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// NewServer initializes the MCP server with service reference and I/O streams.
func NewServer(svc *service.AggregatorService, in io.Reader, out io.Writer) *Server {
	s := &Server{
		svc:    svc,
		in:     in,
		out:    out,
		errLog: log.New(os.Stderr, "[organizer-mcp] ", log.LstdFlags),
	}
	s.initTools()
	return s
}

func (s *Server) initTools() {
	s.tools = []Tool{
		{
			Name:        "organizer_list_items",
			Description: "List and filter AI skills and MCP servers cataloged across all local ecosystems (Grok, Hermes, OpenCode, Cursor, etc.). Supports filtering by type, provider, category, subCategory, classification status, and search text.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"skill", "mcp"},
						"description": "Filter by item type (skill or mcp)",
					},
					"provider": map[string]any{
						"type":        "string",
						"description": "Filter by AI client provider (e.g. grok, hermes, cursor, claude, opencode, kimi, pi, global)",
					},
					"origin": map[string]any{
						"type":        "string",
						"description": "Filter by origin (builtin, installed, custom)",
					},
					"category": map[string]any{
						"type":        "string",
						"description": "Filter by taxonomy category (e.g. 'Bug Bounty & Security', 'Gentle AI & SDD', 'Frontend & UI/UX')",
					},
					"subCategory": map[string]any{
						"type":        "string",
						"description": "Filter by taxonomy subcategory (e.g. 'Web Vulnerabilities', 'Spec-Driven Development')",
					},
					"family": map[string]any{
						"type":        "string",
						"description": "Filter by family or parent skill slug (e.g. 'vue', 'hunt', 'sdd', 'game', 'golang')",
					},
					"isClassified": map[string]any{
						"type":        "boolean",
						"description": "Filter by classification state: true for classified, false for pending triage",
					},
					"search": map[string]any{
						"type":        "string",
						"description": "Keyword search across name, description, command, invocation, source path",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of items to return (default 50)",
					},
				},
			},
		},
		{
			Name:        "organizer_get_item",
			Description: "Get complete inspection details, command, arguments, environment keys, and invocation guidelines for a specific item by its unique ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "The unique item ID (e.g. 'grok:mcp:obsidian', 'global:skill:hunt-sqli')",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "organizer_get_unclassified",
			Description: "List all tools and skills that currently lack taxonomy classification and are pending triage (Category == 'Unclassified'). Useful for catalog enrichment.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum items to return (default 50)",
					},
				},
			},
		},
		{
			Name:        "organizer_add_taxonomy_rule",
			Description: "Add a new declarative taxonomy classification rule (with glob/regex patterns) and immediately re-classify all items in the catalog.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{
						"type":        "string",
						"description": "Target category (e.g. 'Bug Bounty & Security', 'Frontend & UI/UX', 'Cloud & Integrations')",
					},
					"subCategory": map[string]any{
						"type":        "string",
						"description": "Target subcategory (e.g. 'Web Vulnerabilities', 'Design Systems & Tokens')",
					},
					"pattern": map[string]any{
						"type":        "string",
						"description": "Pattern to match (e.g. 'hunt-*', 'sqli', 'brandkit*')",
					},
					"patterns": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Optional array of pattern strings",
					},
				},
				"required": []string{"category", "subCategory"},
			},
		},
		{
			Name:        "organizer_get_stats",
			Description: "Retrieve aggregate statistics across all cataloged items: total count, breakdown by type, provider, origin, classified count, unclassified count, and taxonomy distribution.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "organizer_refresh",
			Description: "Re-scan all local AI client configuration directories and return fresh catalog statistics.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}

// Serve reads JSON-RPC requests from the input reader and dispatches them until EOF or ctx cancel.
func (s *Server) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	// Allow lines up to 10MB
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	s.errLog.Println("Organizer MCP stdio server started and listening...")

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req RPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.errLog.Printf("Failed to parse JSON-RPC request: %v", err)
			s.sendError(nil, -32700, "Parse error", nil)
			continue
		}

		s.handleRequest(ctx, req)
	}

	if err := scanner.Err(); err != nil {
		s.errLog.Printf("Scanner read error: %v", err)
		return err
	}

	s.errLog.Println("Organizer MCP server stdio reached EOF, exiting.")
	return nil
}

func (s *Server) handleRequest(ctx context.Context, req RPCRequest) {
	isNotification := len(req.ID) == 0 || string(req.ID) == "null"

	switch req.Method {
	case "initialize":
		s.sendResult(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "organizer-mcp",
				"version": "1.0.0",
			},
		})

	case "notifications/initialized", "initialized":
		// Notification - no reply needed
		return

	case "ping":
		if !isNotification {
			s.sendResult(req.ID, map[string]any{})
		}

	case "tools/list":
		s.sendResult(req.ID, map[string]any{
			"tools": s.tools,
		})

	case "tools/call":
		res, err := s.handleToolCall(ctx, req.Params)
		if err != nil {
			s.sendError(req.ID, -32603, err.Error(), nil)
			return
		}
		s.sendResult(req.ID, res)

	default:
		if !isNotification {
			s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
		}
	}
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolCall(ctx context.Context, paramsRaw json.RawMessage) (ToolCallResult, error) {
	var params toolCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Invalid tools/call params: %v", err)}},
			IsError: true,
		}, nil
	}

	switch params.Name {
	case "organizer_list_items":
		return s.toolListItems(params.Arguments)
	case "organizer_get_item":
		return s.toolGetItem(params.Arguments)
	case "organizer_get_unclassified":
		return s.toolGetUnclassified(params.Arguments)
	case "organizer_add_taxonomy_rule":
		return s.toolAddTaxonomyRule(params.Arguments)
	case "organizer_get_stats":
		return s.toolGetStats()
	case "organizer_refresh":
		return s.toolRefresh(ctx)
	default:
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", params.Name)}},
			IsError: true,
		}, nil
	}
}

func (s *Server) toolListItems(argsRaw json.RawMessage) (ToolCallResult, error) {
	type listArgs struct {
		Type         string `json:"type"`
		Provider     string `json:"provider"`
		Origin       string `json:"origin"`
		Category     string `json:"category"`
		SubCategory  string `json:"subCategory"`
		Family       string `json:"family"`
		IsClassified *bool  `json:"isClassified"`
		Search       string `json:"search"`
		Limit        int    `json:"limit"`
	}

	var args listArgs
	if len(argsRaw) > 0 {
		_ = json.Unmarshal(argsRaw, &args)
	}

	items := s.svc.GetItems(
		domain.ItemType(args.Type),
		domain.Provider(args.Provider),
		args.Origin,
		args.Category,
		args.SubCategory,
		args.Family,
		args.IsClassified,
		args.Search,
	)

	limit := args.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	totalFound := len(items)
	if len(items) > limit {
		items = items[:limit]
	}

	payload := map[string]any{
		"totalMatches": totalFound,
		"returned":     len(items),
		"items":        items,
	}

	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) toolGetItem(argsRaw json.RawMessage) (ToolCallResult, error) {
	type getArgs struct {
		ID string `json:"id"`
	}
	var args getArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil || args.ID == "" {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: "Missing required 'id' argument"}},
			IsError: true,
		}, nil
	}

	item, found := s.svc.GetItemByID(args.ID)
	if !found {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Item not found with id: %s", args.ID)}},
			IsError: true,
		}, nil
	}

	bytes, _ := json.MarshalIndent(item, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) toolGetUnclassified(argsRaw json.RawMessage) (ToolCallResult, error) {
	type unclassArgs struct {
		Limit int `json:"limit"`
	}
	var args unclassArgs
	if len(argsRaw) > 0 {
		_ = json.Unmarshal(argsRaw, &args)
	}

	isClassified := false
	items := s.svc.GetItems("", "", "", "", "", "", &isClassified, "")

	limit := args.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	totalFound := len(items)
	if len(items) > limit {
		items = items[:limit]
	}

	payload := map[string]any{
		"unclassifiedTotal": totalFound,
		"returned":          len(items),
		"items":             items,
	}

	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) toolAddTaxonomyRule(argsRaw json.RawMessage) (ToolCallResult, error) {
	type addArgs struct {
		Category    string   `json:"category"`
		SubCategory string   `json:"subCategory"`
		Pattern     string   `json:"pattern"`
		Patterns    []string `json:"patterns"`
	}
	var args addArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Invalid arguments: %v", err)}},
			IsError: true,
		}, nil
	}

	patterns := args.Patterns
	if args.Pattern != "" {
		patterns = append(patterns, args.Pattern)
	}

	tax := s.svc.GetTaxonomy()
	if tax == nil {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: "Taxonomy engine not available"}},
			IsError: true,
		}, nil
	}

	if err := tax.AddRule(args.Category, args.SubCategory, patterns); err != nil {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Failed to add rule: %v", err)}},
			IsError: true,
		}, nil
	}

	s.svc.ReclassifyAll()
	stats := s.svc.GetStats()

	payload := map[string]any{
		"status":            "ok",
		"category":          args.Category,
		"subCategory":       args.SubCategory,
		"patterns":          patterns,
		"classifiedCount":   stats.ClassifiedCount,
		"unclassifiedCount": stats.UnclassifiedCount,
	}

	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) toolGetStats() (ToolCallResult, error) {
	stats := s.svc.GetStats()
	bytes, _ := json.MarshalIndent(stats, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) toolRefresh(ctx context.Context) (ToolCallResult, error) {
	items, err := s.svc.ScanAll(ctx)
	if err != nil {
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Scan error: %v", err)}},
			IsError: true,
		}, nil
	}

	stats := s.svc.GetStats()
	payload := map[string]any{
		"status":     "ok",
		"itemsCount": len(items),
		"stats":      stats,
	}

	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return ToolCallResult{
		Content: []TextContent{{Type: "text", Text: string(bytes)}},
	}, nil
}

func (s *Server) sendResult(id json.RawMessage, result any) {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	resp := RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		s.errLog.Printf("Failed to marshal JSON-RPC response: %v", err)
		return
	}
	_, _ = s.out.Write(append(bytes, '\n'))
}

func (s *Server) sendError(id json.RawMessage, code int, message string, data any) {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	resp := RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		s.errLog.Printf("Failed to marshal JSON-RPC error response: %v", err)
		return
	}
	_, _ = s.out.Write(append(bytes, '\n'))
}
