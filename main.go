package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xpzouying/xiaohongshu-mcp/internal/server"
)

var (
	// Version is set at build time
	Version = "dev"
	// BuildTime is set at build time
	BuildTime = "unknown"
)

func main() {
	var (
		port    = flag.Int("port", 8080, "HTTP server port")
		version = flag.Bool("version", false, "Print version information")
		debug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	if *version {
		fmt.Printf("xiaohongshu-mcp version %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Configure logger
	logger := log.New(os.Stdout, "[xiaohongshu-mcp] ", log.LstdFlags)
	if *debug {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
		logger.Println("Debug mode enabled")
	}

	logger.Printf("Starting xiaohongshu-mcp server v%s on port %d", Version, *port)

	// Create MCP server
	cfg := server.Config{
		Port:    *port,
		Debug:   *debug,
		Logger:  logger,
	}

	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	logger.Println("Shutting down server...")
	cancel()

	if err := srv.Stop(); err != nil {
		logger.Printf("Error during shutdown: %v", err)
	}

	logger.Println("Server stopped")
}
