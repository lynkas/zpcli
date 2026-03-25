package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"zpcli/internal/domain"
	"zpcli/store"
)

func TestHandleToolCallSearchVideosUsesProvidedPage(t *testing.T) {
	originalLoadStore := mcpLoadStore
	originalSearchVideos := mcpSearchVideos
	t.Cleanup(func() {
		mcpLoadStore = originalLoadStore
		mcpSearchVideos = originalSearchVideos
	})

	mcpLoadStore = func() (*store.StoreData, error) {
		return &store.StoreData{
			Version: store.CurrentVersion,
			Series: []*store.Series{
				{
					Domains: []*store.Domain{
						{URL: "example.com", FailureScore: 0},
					},
				},
			},
		}, nil
	}

	var gotKeyword string
	var gotPage int
	mcpSearchVideos = func(_ context.Context, data *store.StoreData, keyword string, page int) ([]domain.SearchResult, error) {
		gotKeyword = keyword
		gotPage = page
		return []domain.SearchResult{
			{
				SeriesIndex: 0,
				DomainIndex: 0,
				DomainURL:   data.Series[0].Domains[0].URL,
				Items: []domain.SearchItem{
					{VodID: 123, VodName: "Movie Title", TypeName: "Movie", VodTime: "2026-03-20 10:00:00", VodRemarks: "HD"},
				},
			},
		}, nil
	}

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name":"search_videos",
			"arguments":{"keyword":"movie","page":2}
		}`),
	}

	var out bytes.Buffer
	handleToolCall(&out, req)

	if gotKeyword != "movie" {
		t.Fatalf("expected keyword movie, got %q", gotKeyword)
	}
	if gotPage != 2 {
		t.Fatalf("expected page 2, got %d", gotPage)
	}
	if !strings.Contains(out.String(), "Movie Title") {
		t.Fatalf("expected rendered search result, got %s", out.String())
	}
}

func TestHandleToolCallSearchVideosNoResultsGuidance(t *testing.T) {
	originalLoadStore := mcpLoadStore
	originalSearchVideos := mcpSearchVideos
	t.Cleanup(func() {
		mcpLoadStore = originalLoadStore
		mcpSearchVideos = originalSearchVideos
	})

	mcpLoadStore = func() (*store.StoreData, error) {
		return &store.StoreData{
			Version: store.CurrentVersion,
			Series: []*store.Series{
				{
					Domains: []*store.Domain{
						{URL: "example.com", FailureScore: 0},
					},
				},
			},
		}, nil
	}

	mcpSearchVideos = func(_ context.Context, data *store.StoreData, keyword string, page int) ([]domain.SearchResult, error) {
		return []domain.SearchResult{
			{
				SeriesIndex: 0,
				DomainIndex: 0,
				DomainURL:   data.Series[0].Domains[0].URL,
			},
		}, nil
	}

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name":"search_videos",
			"arguments":{"keyword":"movie","page":3}
		}`),
	}

	var out bytes.Buffer
	handleToolCall(&out, req)

	response := out.String()
	if !strings.Contains(response, "No results found on page 3.") {
		t.Fatalf("expected page-aware no-results message, got %s", response)
	}
	if !strings.Contains(response, "Try the next page with the same keyword") {
		t.Fatalf("expected paging guidance, got %s", response)
	}
	if !strings.Contains(response, "Do not expand a short core keyword") {
		t.Fatalf("expected keyword guidance, got %s", response)
	}
}

func TestSearchToolSchemaIncludesPage(t *testing.T) {
	schema := searchToolSchema()
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected schema properties, got %#v", schema["properties"])
	}

	page, ok := properties["page"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected page property, got %#v", properties["page"])
	}
	if page["minimum"] != 1 {
		t.Fatalf("expected page minimum 1, got %#v", page["minimum"])
	}

	description, _ := page["description"].(string)
	if !strings.Contains(description, "same keyword") {
		t.Fatalf("expected paging guidance in page description, got %q", description)
	}
}
