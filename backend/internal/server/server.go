package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"organizer/backend/internal/domain"
	"organizer/backend/internal/service"
)

// Server encapsulates the HTTP API and routing.
type Server struct {
	svc *service.AggregatorService
	mux *http.ServeMux
}

// NewServer initializes HTTP routes and handlers.
func NewServer(svc *service.AggregatorService) *Server {
	s := &Server{
		svc: svc,
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

// Handler returns the HTTP handler wrapped with CORS middleware.
func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/items", s.handleGetItems)
	s.mux.HandleFunc("GET /api/stats", s.handleGetStats)
	s.mux.HandleFunc("POST /api/refresh", s.handleRefresh)
	s.mux.HandleFunc("GET /api/taxonomy", s.handleGetTaxonomy)
	s.mux.HandleFunc("POST /api/taxonomy", s.handleAddTaxonomyRule)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	itemType := domain.ItemType(q.Get("type"))
	provider := domain.Provider(q.Get("provider"))
	origin := q.Get("origin")
	category := q.Get("category")
	subCategory := q.Get("subCategory")
	family := q.Get("family")
	search := q.Get("search")

	var isClassified *bool
	isClassifiedStr := strings.TrimSpace(strings.ToLower(q.Get("isClassified")))
	if isClassifiedStr == "true" || isClassifiedStr == "1" {
		t := true
		isClassified = &t
	} else if isClassifiedStr == "false" || isClassifiedStr == "0" {
		f := false
		isClassified = &f
	}

	items := s.svc.GetItems(itemType, provider, origin, category, subCategory, family, isClassified, search)
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.svc.GetStats()
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ScanAll(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"itemsCount": len(items),
	})
}

func (s *Server) handleGetTaxonomy(w http.ResponseWriter, r *http.Request) {
	tax := s.svc.GetTaxonomy()
	if tax == nil {
		respondJSON(w, http.StatusOK, map[string]any{
			"rules":      []any{},
			"categories": []any{},
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"rules":      tax.GetRules(),
		"categories": tax.GetCategories(),
	})
}

type addRuleRequest struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"subCategory"`
	Pattern     string   `json:"pattern,omitempty"`
	Patterns    []string `json:"patterns,omitempty"`
}

func (s *Server) handleAddTaxonomyRule(w http.ResponseWriter, r *http.Request) {
	var req addRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	patterns := req.Patterns
	if req.Pattern != "" {
		patterns = append(patterns, req.Pattern)
	}

	tax := s.svc.GetTaxonomy()
	if tax == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "taxonomy engine not initialized"})
		return
	}

	if err := tax.AddRule(req.Category, req.SubCategory, patterns); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Reclassify all cached items so the new rule takes effect immediately
	s.svc.ReclassifyAll()

	respondJSON(w, http.StatusCreated, map[string]any{
		"status":      "ok",
		"category":    req.Category,
		"subCategory": req.SubCategory,
		"patterns":    patterns,
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
