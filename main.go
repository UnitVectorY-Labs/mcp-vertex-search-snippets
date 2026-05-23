package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/UnitVectorY-Labs/mcp-vertex-search-snippets/internal/vertex"
)

var Version = "dev"
const projectName = "mcp-vertex-search-snippets"

func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	var httpAddr string
	var cfgFlag string
	var dbg bool
	var showVersion bool

	flag.StringVar(&httpAddr, "http", "", "run HTTP transport on port (e.g., 8080); defaults to stdio")
	flag.StringVar(&cfgFlag, "vertexConfig", "", "path to the configuration YAML file(overrides VERTEX_CONFIG)")
	flag.BoolVar(&dbg, "vertexDebug", false, "enable debug logging (overrides VERTEX_DEBUG)")
	flag.BoolVar(&showVersion, "version", false, "show version information")
	for _, arg := range os.Args[1:] {
		if arg == "-version" || arg == "--version" || strings.HasPrefix(arg, "-version=") || strings.HasPrefix(arg, "--version=") {
			fmt.Printf("%s version %s (%s, %s/%s)\n", projectName, Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
			os.Exit(0)
		}
	}
	flag.Parse()
	if showVersion {
		fmt.Printf("%s version %s (%s, %s/%s)\n", projectName, Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

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
