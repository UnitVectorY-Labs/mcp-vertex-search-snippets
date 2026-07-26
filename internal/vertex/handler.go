package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/oauth2/google"
)

func boolPtr(b bool) *bool {
	return &b
}

type SearchInput struct {
	Query                    string `json:"query" jsonschema:"Search text"`
	MaxExtractiveSegmentCount *int  `json:"maxExtractiveSegmentCount,omitempty" jsonschema:"Maximum number of extractive segments to return (default: 1)"`
}

type SearchOutput struct {
	Text string `json:"text" jsonschema:"search results as plain text"`
}

func CreateMCPServer(app *AppConfig, version string) (*mcp.Server, error) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "mcp-vertex-search-snippets", Version: version}, nil)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Search for relevant documents based on the provided query.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Search",
			ReadOnlyHint:    true,
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(true),
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (
		*mcp.CallToolResult, SearchOutput, error,
	) {
		q := strings.TrimSpace(input.Query)
		if q == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "missing required argument: query"}},
				IsError: true,
			}, SearchOutput{}, nil
		}

		maxSegments := 1
		if input.MaxExtractiveSegmentCount != nil {
			maxSegments = *input.MaxExtractiveSegmentCount
		}

		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to find default credentials: %v", err)}},
				IsError: true,
			}, SearchOutput{}, nil
		}

		tokenSource := creds.TokenSource
		token, err := tokenSource.Token()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to acquire access token: %v", err)}},
				IsError: true,
			}, SearchOutput{}, nil
		}

		bearerToken := fmt.Sprintf("Bearer %s", token.AccessToken)

		body := searchRequest{
			Query: q,
			ContentSearchSpec: &contentSearchSpec{
				ExtractiveContentSpec: &extractiveContentSpec{
					MaxExtractiveSegmentCount: maxSegments,
				},
			},
		}
		raw, status, err := PostSearch(app.Config.URL(), bearerToken, body, app.IsDebug)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Vertex search failed: %v", err)}},
				IsError: true,
			}, SearchOutput{}, nil
		}
		if status < 200 || status >= 300 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Vertex search HTTP %d: %s", status, string(raw))}},
				IsError: true,
			}, SearchOutput{}, nil
		}

		text := extractText(raw)
		if strings.TrimSpace(text) == "" {
			return nil, SearchOutput{Text: "No content found for the query."}, nil
		}

		return nil, SearchOutput{Text: text}, nil
	})

	return srv, nil
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
		if len(ds.ExtractiveSegments) > 0 {
			for _, seg := range ds.ExtractiveSegments {
				if s := strings.TrimSpace(seg.Content); s != "" {
					parts = append(parts, s)
				}
			}
			continue
		}
		if len(ds.Snippets) > 0 {
			for _, sn := range ds.Snippets {
				if s := strings.TrimSpace(sn.Snippet); s != "" {
					parts = append(parts, s)
				}
			}
			continue
		}
		if ds.Title != "" || ds.Link != "" {
			parts = append(parts, strings.TrimSpace(strings.TrimSpace(ds.Title)+" - "+strings.TrimSpace(ds.Link)))
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
