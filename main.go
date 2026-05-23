package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/UnitVectorY-Labs/mcp-vertex-search-snippets/internal/vertex"
)

var Version = "dev"

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

const projectName = "mcp-vertex-search-snippets"

func versionString() string {
	version := Version
	if semverRe.MatchString(version) && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return fmt.Sprintf("%s version %s (%s, %s/%s)", projectName, version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func printVersionAndExit() {
	fmt.Println(versionString())
	os.Exit(0)
}

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
	flag.Parse()
	if showVersion {
		printVersionAndExit()
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
