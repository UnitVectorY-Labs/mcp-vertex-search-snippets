package vertex

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServeOptions struct {
	HTTPAddr string
	IsDebug  bool
}

func Serve(srv *mcp.Server, opts ServeOptions) error {
	if opts.HTTPAddr != "" {
		return serveHTTP(srv, opts.HTTPAddr, opts.IsDebug)
	}
	return serveStdio(srv)
}

func serveHTTP(srv *mcp.Server, httpAddr string, debug bool) error {
	if debug {
		fmt.Printf("Starting MCP server (HTTP) on %s\n", httpAddr)
	}
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return srv
	}, nil)
	if debug {
		fmt.Printf("Endpoint: http://localhost:%s/mcp\n", httpAddr)
	}
	return http.ListenAndServe(":"+httpAddr, handler)
}

func serveStdio(srv *mcp.Server) error {
	return srv.Run(context.Background(), &mcp.StdioTransport{})
}
