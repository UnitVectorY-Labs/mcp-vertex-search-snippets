package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

type searchRequest struct {
	Query             string             `json:"query"`
	ContentSearchSpec *contentSearchSpec `json:"contentSearchSpec,omitempty"`
	// Additional parameters can be added here to further customize the query
}

type contentSearchSpec struct {
	ExtractiveContentSpec *extractiveContentSpec `json:"extractiveContentSpec,omitempty"`
}

type extractiveContentSpec struct {
	MaxExtractiveSegmentCount int `json:"maxExtractiveSegmentCount,omitempty"`
}

func PostSearch(url, token string, reqBody searchRequest, debug bool) ([]byte, int, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	if debug {
		log.Println("--- Vertex Search Request ---")
		if dump, err := httputil.DumpRequestOut(req, true); err == nil {
			log.Printf("%s\n", dump)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if debug {
		log.Println("--- Vertex Search Response ---")
		log.Printf("Status Code: %d\n", resp.StatusCode)
		var pretty bytes.Buffer
		if json.Indent(&pretty, body, "", "  ") == nil {
			log.Printf("Body:\n%s\n", pretty.String())
		} else {
			log.Printf("Body (raw): %s\n", body)
		}
	}

	return body, resp.StatusCode, nil
}
