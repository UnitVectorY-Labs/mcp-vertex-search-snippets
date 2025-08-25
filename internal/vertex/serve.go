package vertex

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
)

type ServeOptions struct {
	HTTPAddr string
	IsDebug  bool
}

func Serve(srv *server.MCPServer, opts ServeOptions) error {
	if opts.HTTPAddr != "" {
		return serveHTTP(srv, opts.HTTPAddr, opts.IsDebug)
	}
	return serveStdio(srv)
}

func serveHTTP(srv *server.MCPServer, httpAddr string, debug bool) error {
	if debug {
		fmt.Printf("Starting MCP server (HTTP) on %s\n", httpAddr)
	}
	streamSrv := server.NewStreamableHTTPServer(
		srv,
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			if auth := r.Header.Get("Authorization"); auth != "" {
				ctx = context.WithValue(ctx, ctxAuthKey{}, auth)
			}
			return ctx
		}),
	)
	if debug {
		fmt.Printf("Endpoint: http://localhost:%s/mcp\n", httpAddr)
	}
	return streamSrv.Start(":" + httpAddr)
}

func serveStdio(srv *server.MCPServer) error {
	return server.ServeStdio(srv)
}
