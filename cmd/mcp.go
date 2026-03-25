package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"zpcli/internal/buildinfo"
	"zpcli/internal/domain"
	"zpcli/internal/logx"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var (
	mcpPort int
)

var (
	mcpLoadStore    = store.Load
	mcpSearchVideos = func(ctx context.Context, data *store.StoreData, keyword string, page int) ([]domain.SearchResult, error) {
		return service.NewSearchService(nil).Search(ctx, data, keyword, 3, page)
	}
)

// JSON-RPC types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP types
type InitializeResult struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ServerInfo      ServerInfo  `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func searchToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"keyword": map[string]interface{}{
				"type":        "string",
				"description": "Required. Search keyword. Keep the core keyword unchanged when paging; if a short keyword finds nothing, adding extra surrounding words usually will not help.",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Optional. Result page number, starting at 1. If page 1 does not contain the expected result, try later pages with the same keyword before changing the query.",
				"minimum":     1,
			},
		},
		"required": []string{"keyword"},
	}
}

func parsePositiveIntArgument(args map[string]interface{}, key string, defaultValue int) int {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue
	}

	switch v := raw.(type) {
	case float64:
		if v >= 1 && float64(int(v)) == v {
			return int(v)
		}
	case int:
		if v >= 1 {
			return v
		}
	case int32:
		if v >= 1 {
			return int(v)
		}
	case int64:
		if v >= 1 {
			return int(v)
		}
	case json.Number:
		if parsed, err := strconv.Atoi(v.String()); err == nil && parsed >= 1 {
			return parsed
		}
	case string:
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 1 {
			return parsed
		}
	}

	return defaultValue
}

func detailToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"site_id": map[string]interface{}{
				"type":        "string",
				"description": "Required. Configured site ID, such as 1.1.",
			},
			"vod_id": map[string]interface{}{
				"type":        "string",
				"description": "Required. Video ID on the selected site.",
			},
			"episode": map[string]interface{}{
				"type":        "string",
				"description": "Optional. When provided, return the matching episode URL instead of full detail text.",
			},
		},
		"required": []string{"site_id", "vod_id"},
	}
}

func listSitesToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func addSiteToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"domain": map[string]interface{}{
				"type":        "string",
				"description": "Required. Bare host or full endpoint URL to add.",
			},
			"series_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional. Existing series ID to append to. If omitted, a new series is created.",
			},
		},
		"required": []string{"domain"},
	}
}

func removeSiteToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Required. Series ID like '1' or domain ID like '1.1'.",
			},
		},
		"required": []string{"id"},
	}
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server to provide video CMS tools to LLMs",
	Long: `Start the MCP server.

Supported forms:
  1. ` + "`zpcli mcp`" + `
     Required:
       - no positional arguments
     Optional:
       - none
     Behavior:
       - starts the MCP server over stdio

  2. ` + "`zpcli mcp --port <port>`" + `
     Required:
       - no positional arguments
     Optional:
       - ` + "`--port`" + `
     Behavior:
       - starts the MCP server as an SSE / HTTP service on the given port`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if mcpPort > 0 {
			serveSSE(mcpPort)
		} else {
			serveStdio()
		}
	},
}

func serveStdio() {
	logger := logx.Logger("cmd.mcp")
	logger.Info("mcp stdio server start")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		logger.Info("mcp stdio input", "raw", string(line))
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			logger.Error("mcp stdio parse error", "error", err, "raw", string(line))
			sendError(os.Stdout, nil, -32700, "Parse error", nil)
			continue
		}
		handleRequest(os.Stdout, req)
	}
}

type sseSession struct {
	id     string
	msgs   chan []byte
	writer io.Writer
}

var (
	sessions = make(map[string]*sseSession)
	sessMu   sync.Mutex
)

func serveSSE(port int) {
	logger := logx.Logger("cmd.mcp")
	logger.Info("mcp sse server start", "port", port)

	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("mcp sse connect", "remote_addr", r.RemoteAddr)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
		s := &sseSession{
			id:   sessionID,
			msgs: make(chan []byte, 10),
		}

		sessMu.Lock()
		sessions[sessionID] = s
		sessMu.Unlock()

		defer func() {
			sessMu.Lock()
			delete(sessions, sessionID)
			sessMu.Unlock()
			logger.Info("mcp sse disconnect", "session_id", sessionID)
		}()

		// Send endpoint info
		fmt.Fprintf(w, "event: endpoint\ndata: /messages?sessionId=%s\n\n", sessionID)
		w.(http.Flusher).Flush()

		for msg := range s.msgs {
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(msg))
			w.(http.Flusher).Flush()
		}
	})

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("mcp sse message request", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.URL.Query().Get("sessionId")
		sessMu.Lock()
		s, ok := sessions[sessionID]
		sessMu.Unlock()

		if !ok {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("mcp sse decode request failed", "session_id", sessionID, "error", err)
			return
		}
		logger.Info("mcp sse decoded request", "session_id", sessionID, "method", req.Method, "id", req.ID, "params", string(req.Params))

		// Wrap the session's channel as a writer
		handleRequest(&chanWriter{s.msgs}, req)
		w.WriteHeader(http.StatusAccepted)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
	}
}

type chanWriter struct {
	ch chan []byte
}

func (cw *chanWriter) Write(p []byte) (n int, err error) {
	cw.ch <- p
	return len(p), nil
}

func handleRequest(w io.Writer, req JSONRPCRequest) {
	logger := logx.Logger("cmd.mcp")
	logger.Info("mcp handle request", "method", req.Method, "id", req.ID, "params", string(req.Params))
	switch req.Method {
	case "initialize":
		sendResponse(w, req.ID, InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "zpcli",
				Version: buildinfo.Version,
			},
		})
	case "notifications/initialized":
	// No response needed for notifications
	case "tools/list":
		sendResponse(w, req.ID, ToolListResult{
			Tools: []Tool{
				{
					Name:        "search_videos",
					Description: "Search videos across configured sites. Required: keyword. Optional: page. If page 1 misses the expected result, try later pages with the same keyword before changing the query. Do not expand a short core keyword with extra words.",
					InputSchema: searchToolSchema(),
				},
				{
					Name:        "get_video_detail",
					Description: "Get video detail. Required: site_id, vod_id. Optional: episode.",
					InputSchema: detailToolSchema(),
				},
				{
					Name:        "list_sites",
					Description: "List all configured series, domain IDs, URLs, and failure counts. No input required.",
					InputSchema: listSitesToolSchema(),
				},
				{
					Name:        "add_site",
					Description: "Add a site domain. Required: domain. Optional: series_id. Without series_id, creates a new series.",
					InputSchema: addSiteToolSchema(),
				},
				{
					Name:        "remove_site",
					Description: "Remove a series or one domain. Required: id, using '1' for a series or '1.1' for a domain.",
					InputSchema: removeSiteToolSchema(),
				},
				{
					Name:        "validate_sites",
					Description: "Validate the current site configuration and report issues. No input required.",
					InputSchema: listSitesToolSchema(),
				},
				{
					Name:        "health_check",
					Description: "Return config health totals, warnings, errors, and config path information. No input required.",
					InputSchema: listSitesToolSchema(),
				},
			},
		})
	case "tools/call":
		handleToolCall(w, req)
	default:
		logger.Warn("mcp unknown method", "method", req.Method, "id", req.ID)
		if req.ID != nil {
			sendError(w, req.ID, -32601, "Method not found", nil)
		}
	}
}

func handleToolCall(w io.Writer, req JSONRPCRequest) {
	logger := logx.Logger("cmd.mcp")
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		logger.Error("mcp tool call decode failed", "id", req.ID, "error", err, "params", string(req.Params))
		sendError(w, req.ID, -32602, "Invalid params", nil)
		return
	}
	if params.Arguments == nil {
		params.Arguments = make(map[string]interface{})
	}
	logger.Info("mcp tool call start", "id", req.ID, "tool", params.Name, "arguments", params.Arguments)

	var result ToolCallResult
	var buf bytes.Buffer
	siteService := service.NewSiteService()

	switch params.Name {
	case "search_videos":
		keyword, _ := params.Arguments["keyword"].(string)
		page := parsePositiveIntArgument(params.Arguments, "page", 1)
		data, err := mcpLoadStore()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		searchResults, err := mcpSearchVideos(context.Background(), data, keyword, page)
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
			break
		}
		if len(searchResults) == 0 {
			result.Content = append(result.Content, Content{Type: "text", Text: "No valid domains to search.\n"})
			break
		}
		hasFailures := false
		for _, searchResult := range searchResults {
			if searchResult.Err != nil {
				data.Series[searchResult.SeriesIndex].Domains[searchResult.DomainIndex].FailureScore++
				hasFailures = true
			}
		}
		if hasFailures {
			data.Save()
		}
		writeSearchResults(&buf, data, searchResults, "time")
		searchText := buf.String()
		if searchText == "No results found.\n" {
			searchText = fmt.Sprintf("No results found on page %d.\nTry the next page with the same keyword before changing the query.\nDo not expand a short core keyword with extra surrounding words.\n", page)
		}
		result.Content = append(result.Content, Content{Type: "text", Text: searchText})
	case "get_video_detail":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		siteID, _ := params.Arguments["site_id"].(string)
		vodID, _ := params.Arguments["vod_id"].(string)
		episode, _ := params.Arguments["episode"].(string)
		detailService := service.NewDetailService(nil)
		detailResult, err := detailService.GetDetail(context.Background(), data, siteID, vodID)
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
			break
		}
		if detailResult.Err != nil {
			data.Series[detailResult.SeriesIndex].Domains[detailResult.DomainIndex].FailureScore++
			data.Save()
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", detailResult.Err)})
			break
		}
		if detailResult.Item == nil {
			result.Content = append(result.Content, Content{Type: "text", Text: "No detail found.\n"})
			break
		}
		if episode != "" {
			writeEpisodeMatch(&buf, detailResult.Item, episode)
			result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
			break
		}
		writeDetailResult(&buf, detailResult.Item, true)
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "list_sites":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		seriesList := siteService.ListSites(data)
		if len(seriesList) == 0 {
			result.Content = append(result.Content, Content{Type: "text", Text: "No series configured.\n"})
			break
		}
		for _, series := range seriesList {
			fmt.Fprintf(&buf, "Series %d:\n", series.SeriesID)
			for _, dom := range series.Domains {
				fmt.Fprintf(&buf, "  [%d.%d] URL: %s [Failures: %d]\n", dom.SeriesID, dom.DomainID, dom.URL, dom.FailureScore)
			}
		}
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "add_site":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		domain, _ := params.Arguments["domain"].(string)
		seriesID, _ := params.Arguments["series_id"].(string)
		var message string
		if seriesID != "" {
			message, err = siteService.AddSite(data, seriesID, domain)
		} else {
			message, err = siteService.AddSite(data, domain)
		}
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
		} else {
			result.Content = append(result.Content, Content{Type: "text", Text: message})
		}
	case "remove_site":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		id, _ := params.Arguments["id"].(string)
		message, err := siteService.RemoveSite(data, id)
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
		} else {
			result.Content = append(result.Content, Content{Type: "text", Text: message})
		}
	case "validate_sites":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		healthService := service.NewHealthService()
		issues := healthService.ValidateStore(data)
		if len(issues) == 0 {
			result.Content = append(result.Content, Content{Type: "text", Text: "Configuration is valid."})
			break
		}
		fmt.Fprintf(&buf, "Found %d issue(s):\n", len(issues))
		for _, issue := range issues {
			label := issue.Scope
			if issue.SiteID != "" {
				label = fmt.Sprintf("%s %s", issue.Scope, issue.SiteID)
			}
			fmt.Fprintf(&buf, "  [%s] %s: %s\n", issue.Level, label, issue.Message)
		}
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "health_check":
		data, err := store.Load()
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error loading store: %v", err)})
			break
		}
		healthService := service.NewHealthService()
		report, err := healthService.BuildHealthReport(data)
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error building health report: %v", err)})
			break
		}
		fmt.Fprintf(&buf, "Config:   %s\n", report.ConfigPath)
		fmt.Fprintf(&buf, "Version:  %d\n", report.Version)
		fmt.Fprintf(&buf, "Series:   %d\n", report.SeriesCount)
		fmt.Fprintf(&buf, "Domains:  %d\n", report.DomainCount)
		fmt.Fprintf(&buf, "Errors:   %d\n", report.InvalidCount)
		fmt.Fprintf(&buf, "Warnings: %d\n", report.WarningCount)
		if len(report.Issues) == 0 {
			fmt.Fprintf(&buf, "\nStatus: healthy\n")
		} else {
			fmt.Fprintf(&buf, "\nIssues:\n")
			for _, issue := range report.Issues {
				label := issue.Scope
				if issue.SiteID != "" {
					label = fmt.Sprintf("%s %s", issue.Scope, issue.SiteID)
				}
				fmt.Fprintf(&buf, "  [%s] %s: %s\n", issue.Level, label, issue.Message)
			}
		}
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	default:
		logger.Warn("mcp unknown tool", "tool", params.Name, "id", req.ID)
		sendError(w, req.ID, -32602, "Unknown tool", nil)
		return
	}

	logger.Info("mcp tool call complete", "id", req.ID, "tool", params.Name, "is_error", result.IsError, "content_count", len(result.Content))
	logger.Debug("mcp tool call result", "id", req.ID, "tool", params.Name, "result", result)
	sendResponse(w, req.ID, result)
}

func sendResponse(w io.Writer, id interface{}, result interface{}) {
	logx.Logger("cmd.mcp").Info("mcp send response", "id", id, "result", result)
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(out))
}

func sendError(w io.Writer, id interface{}, code int, message string, data interface{}) {
	logx.Logger("cmd.mcp").Error("mcp send error", "id", id, "code", code, "message", message, "data", data)
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	out, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(out))
}

func init() {
	mcpCmd.Flags().IntVarP(&mcpPort, "port", "p", 0, "Port to run MCP SSE server on (default 0, uses stdio)")
	rootCmd.AddCommand(mcpCmd)
}
