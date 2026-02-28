package vertex

import (
	"encoding/json"
	"testing"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty response",
			input:    `{}`,
			expected: "",
		},
		{
			name:     "invalid json",
			input:    `{invalid`,
			expected: "",
		},
		{
			name:     "no results",
			input:    `{"results": []}`,
			expected: "",
		},
		{
			name: "extractive segments preferred over snippets",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"title": "Test Doc",
							"link": "https://example.com",
							"snippets": [{"snippet": "snippet text"}],
							"extractive_segments": [{"content": "segment text"}]
						}
					}
				}]
			}`,
			expected: "segment text",
		},
		{
			name: "snippets used when no extractive segments",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"title": "Test Doc",
							"link": "https://example.com",
							"snippets": [{"snippet": "snippet text"}]
						}
					}
				}]
			}`,
			expected: "snippet text",
		},
		{
			name: "title and link fallback",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"title": "Test Doc",
							"link": "https://example.com"
						}
					}
				}]
			}`,
			expected: "Test Doc - https://example.com",
		},
		{
			name: "title only fallback",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"title": "Test Doc"
						}
					}
				}]
			}`,
			expected: "Test Doc -",
		},
		{
			name: "multiple results separated",
			input: `{
				"results": [
					{
						"document": {
							"derivedStructData": {
								"extractive_segments": [{"content": "first segment"}]
							}
						}
					},
					{
						"document": {
							"derivedStructData": {
								"extractive_segments": [{"content": "second segment"}]
							}
						}
					}
				]
			}`,
			expected: "first segment\n\n---\n\nsecond segment",
		},
		{
			name: "multiple extractive segments in one result",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"extractive_segments": [
								{"content": "segment one"},
								{"content": "segment two"}
							]
						}
					}
				}]
			}`,
			expected: "segment one\n\n---\n\nsegment two",
		},
		{
			name: "empty extractive segment content skipped",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"extractive_segments": [
								{"content": ""},
								{"content": "valid segment"}
							]
						}
					}
				}]
			}`,
			expected: "valid segment",
		},
		{
			name: "whitespace-only extractive segments skips to next result",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"extractive_segments": [{"content": "   "}],
							"snippets": [{"snippet": "  real snippet  "}]
						}
					}
				}]
			}`,
			expected: "",
		},
		{
			name: "whitespace-only snippet content skipped",
			input: `{
				"results": [{
					"document": {
						"derivedStructData": {
							"snippets": [
								{"snippet": "   "},
								{"snippet": "valid snippet"}
							]
						}
					}
				}]
			}`,
			expected: "valid snippet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractText([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("extractText() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSearchRequestJSON(t *testing.T) {
	body := searchRequest{
		Query: "test query",
		ContentSearchSpec: &contentSearchSpec{
			SnippetSpec: &snippetSpec{
				ReturnSnippet: true,
			},
			ExtractiveContentSpec: &extractiveContentSpec{
				MaxExtractiveSegmentCount: 3,
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["query"] != "test query" {
		t.Errorf("query = %v, want %q", result["query"], "test query")
	}

	spec, ok := result["contentSearchSpec"].(map[string]interface{})
	if !ok {
		t.Fatal("contentSearchSpec is missing or not an object")
	}

	snippet, ok := spec["snippetSpec"].(map[string]interface{})
	if !ok {
		t.Fatal("snippetSpec is missing or not an object")
	}
	if snippet["returnSnippet"] != true {
		t.Errorf("returnSnippet = %v, want true", snippet["returnSnippet"])
	}

	extractive, ok := spec["extractiveContentSpec"].(map[string]interface{})
	if !ok {
		t.Fatal("extractiveContentSpec is missing or not an object")
	}
	if extractive["maxExtractiveSegmentCount"] != float64(3) {
		t.Errorf("maxExtractiveSegmentCount = %v, want 3", extractive["maxExtractiveSegmentCount"])
	}
}
