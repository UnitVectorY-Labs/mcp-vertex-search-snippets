package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2/google"
)

type ctxAuthKey struct{}

func CreateMCPServer(app *AppConfig, version string) (*server.MCPServer, error) {
	srv := server.NewMCPServer("mcp-vertex-search-snippets", version)

	// One tool: "search"
	opts := []mcp.ToolOption{
		mcp.WithDescription("Search for relevant documents based on the provided query."),
		mcp.WithString("query", mcp.Description("Search text"), mcp.Required()),
		mcp.WithNumber("maxExtractiveSegmentCount", mcp.Description("Maximum number of extractive segments to return (default: 1)")),
		mcp.WithTitleAnnotation("Search"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
	}
	tool := mcp.NewTool("search", opts...)
	srv.AddTool(tool, makeHandler(app))

	return srv, nil
}

func makeHandler(app *AppConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		q, ok := args["query"].(string)
		if !ok || strings.TrimSpace(q) == "" {
			return mcp.NewToolResultError("missing required argument: query"), nil
		}

		// Parse maxExtractiveSegmentCount parameter (optional, defaults to 1)
		maxSegments := 1
		if maxSegmentsArg, exists := args["maxExtractiveSegmentCount"]; exists {
			if maxSegmentsFloat, ok := maxSegmentsArg.(float64); ok {
				maxSegments = int(maxSegmentsFloat)
			} else if maxSegmentsInt, ok := maxSegmentsArg.(int); ok {
				maxSegments = maxSegmentsInt
			}
		}
		if maxSegments < 1 {
			maxSegments = 1
		}

		// Check if an Authorization header was provided via HTTP context
		var bearerToken string
		if auth, ok := ctx.Value(ctxAuthKey{}).(string); ok && auth != "" {
			bearerToken = auth
		} else {
			// Acquire a token using the Google Cloud authentication library
			creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to find default credentials", err), nil
			}

			token, err := creds.TokenSource.Token()
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to acquire access token", err), nil
			}

			bearerToken = fmt.Sprintf("Bearer %s", token.AccessToken)
		}

		// Build search request
		body := searchRequest{
			Query: q,
			ContentSearchSpec: &contentSearchSpec{
				SnippetSpec: &snippetSpec{
					ReturnSnippet: true,
				},
				ExtractiveContentSpec: &extractiveContentSpec{
					MaxExtractiveSegmentCount: maxSegments,
				},
			},
		}
		raw, status, err := PostSearch(app.Config.URL(), bearerToken, body, app.IsDebug)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Vertex search failed", err), nil
		}
		if status < 200 || status >= 300 {
			return mcp.NewToolResultError(fmt.Sprintf("Vertex search HTTP %d: %s", status, string(raw))), nil
		}

		// Build plain-text output from the response:
		text := extractText(raw)
		if strings.TrimSpace(text) == "" {
			// There is no content, return a default message
			return mcp.NewToolResultText("No content found for the query."), nil
		}

		return mcp.NewToolResultText(text), nil
	}
}

type vertexResponse struct {
	Results []struct {
		Document struct {
			Derived struct {
				Title    string `json:"title"`
				Link     string `json:"link"`
				Snippets []struct {
					Snippet string `json:"snippet"`
				} `json:"snippets"`
				ExtractiveSegments []struct {
					Content string `json:"content"`
				} `json:"extractive_segments"`
			} `json:"derivedStructData"`
		} `json:"document"`
	} `json:"results"`
}

func extractText(raw []byte) string {
	var vr vertexResponse
	if err := json.Unmarshal(raw, &vr); err != nil {
		return ""
	}
	var parts []string
	for _, r := range vr.Results {
		ds := r.Document.Derived
		// Prefer extractive segments
		if len(ds.ExtractiveSegments) > 0 {
			for _, seg := range ds.ExtractiveSegments {
				if s := strings.TrimSpace(seg.Content); s != "" {
					parts = append(parts, s)
				}
			}
			continue
		}
		// Then snippets
		if len(ds.Snippets) > 0 {
			for _, sn := range ds.Snippets {
				if s := strings.TrimSpace(sn.Snippet); s != "" {
					parts = append(parts, s)
				}
			}
			continue
		}
		// Then title/link
		if ds.Title != "" || ds.Link != "" {
			parts = append(parts, strings.TrimSpace(strings.TrimSpace(ds.Title)+" - "+strings.TrimSpace(ds.Link)))
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
