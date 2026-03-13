package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	mcpPort int
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

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server to provide video CMS tools to LLMs",
	Run: func(cmd *cobra.Command, args []string) {
		if mcpPort > 0 {
			serveSSE(mcpPort)
		} else {
			serveStdio()
		}
	},
}

func serveStdio() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
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
	fmt.Fprintf(os.Stderr, "Starting MCP SSE server on :%d\n", port)

	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
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
			return
		}

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
	switch req.Method {
	case "initialize":
		sendResponse(w, req.ID, InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "zpcli",
				Version: "1.0.0",
			},
		})
	case "notifications/initialized":
		// No response needed for notifications
	case "tools/list":
		sendResponse(w, req.ID, ToolListResult{
			Tools: []Tool{
				{
					Name:        "search",
					Description: "Search for videos across configured sites",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":        "string",
								"description": "The keyword to search for",
							},
						},
						"required": []string{"keyword"},
					},
				},
				{
					Name:        "get_detail",
					Description: "Get details of a specific video",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"site_id": map[string]interface{}{
								"type":        "string",
								"description": "The site ID (e.g. 2.1)",
							},
							"vod_id": map[string]interface{}{
								"type":        "string",
								"description": "The video ID",
							},
							"episode": map[string]interface{}{
								"type":        "string",
								"description": "Optional specific episode to get link for",
							},
						},
						"required": []string{"site_id", "vod_id"},
					},
				},
				{
					Name:        "list_sites",
					Description: "List all configured sites and their IDs",
					InputSchema: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					Name:        "add_site",
					Description: "Add a new site domain. Can create a new series or add to existing.",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"domain": map[string]interface{}{
								"type":        "string",
								"description": "The domain URL to add",
							},
							"series_id": map[string]interface{}{
								"type":        "string",
								"description": "Optional series ID to add the domain to. If omitted, creates a new series.",
							},
						},
						"required": []string{"domain"},
					},
				},
				{
					Name:        "remove_site",
					Description: "Remove a site or an entire series",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"description": "The ID to remove (e.g., '1.1' for a domain or '1' for a series)",
							},
						},
						"required": []string{"id"},
					},
				},
			},
		})
	case "tools/call":
		handleToolCall(w, req)
	default:
		if req.ID != nil {
			sendError(w, req.ID, -32601, "Method not found", nil)
		}
	}
}

func handleToolCall(w io.Writer, req JSONRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(w, req.ID, -32602, "Invalid params", nil)
		return
	}
	if params.Arguments == nil {
		params.Arguments = make(map[string]interface{})
	}

	var result ToolCallResult
	var buf bytes.Buffer

	switch params.Name {
	case "search":
		keyword, _ := params.Arguments["keyword"].(string)
		ShowSearch(&buf, keyword, 3, 1, "time")
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "get_detail":
		siteID, _ := params.Arguments["site_id"].(string)
		vodID, _ := params.Arguments["vod_id"].(string)
		episode, _ := params.Arguments["episode"].(string)
		if episode != "" {
			ShowDetail(&buf, true, siteID, vodID, episode)
		} else {
			ShowDetail(&buf, true, siteID, vodID)
		}
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "list_sites":
		ShowList(&buf)
		result.Content = append(result.Content, Content{Type: "text", Text: buf.String()})
	case "add_site":
		domain, _ := params.Arguments["domain"].(string)
		seriesID, _ := params.Arguments["series_id"].(string)
		var err error
		if seriesID != "" {
			err = AddSite(seriesID, domain)
		} else {
			err = AddSite(domain)
		}
		if err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
		} else {
			result.Content = append(result.Content, Content{Type: "text", Text: "Site added successfully"})
		}
	case "remove_site":
		id, _ := params.Arguments["id"].(string)
		if err := RemoveSite(id); err != nil {
			result.IsError = true
			result.Content = append(result.Content, Content{Type: "text", Text: fmt.Sprintf("Error: %v", err)})
		} else {
			result.Content = append(result.Content, Content{Type: "text", Text: "Site removed successfully"})
		}
	default:
		sendError(w, req.ID, -32602, "Unknown tool", nil)
		return
	}

	sendResponse(w, req.ID, result)
}

func sendResponse(w io.Writer, id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(out))
}

func sendError(w io.Writer, id interface{}, code int, message string, data interface{}) {
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
