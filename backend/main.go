package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mcpAdapter "organizer/backend/internal/adapter/mcp"
	"organizer/backend/internal/adapter/skills"
	mcpServer "organizer/backend/internal/mcp"
	"organizer/backend/internal/server"
	"organizer/backend/internal/service"
)

func main() {
	isMCP := len(os.Args) > 1 && os.Args[1] == "mcp"
	if isMCP {
		// In stdio MCP mode, all log output MUST go to stderr, never stdout!
		log.SetOutput(os.Stderr)
	}

	log.Println("Starting Organizer AI Scanner Backend...")

	// 1. Initialize client adapters
	skillsScanner := skills.NewScanner()
	standardMCPScanner := mcpAdapter.NewStandardJSONScanner()
	openCodeMCPScanner := mcpAdapter.NewOpenCodeScanner()
	hermesMCPScanner := mcpAdapter.NewHermesScanner()
	grokMCPScanner := mcpAdapter.NewGrokScanner()

	// 2. Initialize Aggregator Service (includes Declarative Taxonomy Engine)
	aggregator := service.NewAggregatorService(
		skillsScanner,
		standardMCPScanner,
		openCodeMCPScanner,
		hermesMCPScanner,
		grokMCPScanner,
	)

	// 3. Perform initial scan
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := aggregator.ScanAll(ctx)
	if err != nil {
		log.Printf("Warning: initial scan had errors: %v", err)
	}

	stats := aggregator.GetStats()
	log.Printf("Initial scan completed: %d total items discovered (%d classified, %d unclassified)",
		stats.Total, stats.ClassifiedCount, stats.UnclassifiedCount)
	for t, count := range stats.ByType {
		log.Printf("  - Type [%s]: %d", t, count)
	}
	for p, count := range stats.ByProvider {
		log.Printf("  - Provider [%s]: %d", p, count)
	}
	for o, count := range stats.ByOrigin {
		log.Printf("  - Origin [%s]: %d", o, count)
	}
	for c, count := range stats.ByCategory {
		log.Printf("  - Category [%s]: %d", c, count)
	}

	// 4. Branch execution: Stdio MCP Server vs HTTP Server
	if isMCP {
		log.Println("Running in MCP stdio mode (compatible with Anthropic/Cursor/Antigravity)...")
		srv := mcpServer.NewServer(aggregator, os.Stdin, os.Stdout)
		mcpCtx, mcpCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer mcpCancel()

		if err := srv.Serve(mcpCtx); err != nil && err != context.Canceled {
			log.Fatalf("MCP server error: %v", err)
		}
		return
	}

	// 5. Configure HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	apiServer := server.NewServer(aggregator)
	httpServer := &http.Server{
		Addr:         port,
		Handler:      apiServer.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Run HTTP server in goroutine
	go func() {
		log.Printf("Organizer HTTP server listening on http://localhost%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 7. Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	fmt.Println("Server exited cleanly.")
}
