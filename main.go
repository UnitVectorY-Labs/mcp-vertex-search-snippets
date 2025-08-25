package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/UnitVectorY-Labs/mcp-vertex-search-snippets/internal/vertex"
)

var Version = "dev"

func main() {
	var httpAddr string
	var cfgFlag string
	var dbg bool

	flag.StringVar(&httpAddr, "http", "", "run HTTP transport on port (e.g., 8080); defaults to stdio")
	flag.StringVar(&cfgFlag, "vertexConfig", "", "path to the folder containing vertex.yaml (overrides VERTEX_CONFIG)")
	flag.BoolVar(&dbg, "vertexDebug", false, "enable debug logging (overrides VERTEX_DEBUG)")
	flag.Parse()

	app, err := vertex.LoadAppConfig(cfgFlag, dbg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if app.IsDebug {
		log.SetOutput(os.Stderr)
		log.Println("Debug mode enabled.")
	} else {
		log.SetOutput(io.Discard)
	}

	srv, err := vertex.CreateMCPServer(app, Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating MCP server: %v\n", err)
		os.Exit(1)
	}

	if err := vertex.Serve(srv, vertex.ServeOptions{HTTPAddr: httpAddr, IsDebug: app.IsDebug}); err != nil {
		log.Fatalf("Fatal: %v\n", err)
	}
}
