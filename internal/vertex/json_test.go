package vertex

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostSearchJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		want  string
	}{
		{"zero omitted", 0, `{"query":"test","contentSearchSpec":{"extractiveContentSpec":{}}}`},
		{"count included", 2, `{"query":"test","contentSearchSpec":{"extractiveContentSpec":{"maxExtractiveSegmentCount":2}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					return
				}
				if string(body) != tc.want {
					t.Errorf("request = %s, want %s", body, tc.want)
				}
				_, _ = w.Write([]byte(`{"results":[]}`))
			}))
			defer server.Close()
			body, status, err := PostSearch(server.URL, "", searchRequest{Query: "test", ContentSearchSpec: &contentSearchSpec{ExtractiveContentSpec: &extractiveContentSpec{MaxExtractiveSegmentCount: tc.count}}}, true)
			if err != nil || status != http.StatusOK || string(body) != `{"results":[]}` {
				t.Fatalf("PostSearch = %s, %d, %v", body, status, err)
			}
		})
	}
}

func TestExtractTextJSON(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"segment", `{"results":[{"document":{"derivedStructData":{"extractive_segments":[{"content":" hello "}]}}}]}`, "hello"},
		{"snippet", `{"results":[{"document":{"derivedStructData":{"snippets":[{"snippet":"hello"}]}}}]}`, "hello"},
		{"duplicate names", `{"results":[],"results":[{"document":{"derivedStructData":{"title":"bad"}}}]}`, ""},
		{"invalid UTF-8", "{\"results\":[{\"document\":{\"derivedStructData\":{\"title\":\"\xff\"}}}]}", ""},
		{"case sensitive", `{"Results":[{"document":{"derivedStructData":{"title":"bad"}}}]}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractText([]byte(tc.input)); got != tc.want {
				t.Errorf("extractText(%q) = %q, want %q", strings.ToValidUTF8(tc.input, "?"), got, tc.want)
			}
		})
	}
}
