package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		serveMCP()
	},
}

func serveMCP() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(nil, -32700, "Parse error", nil)
			continue
		}
		handleRequest(req)
	}
}

func handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(req.ID, InitializeResult{
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
		sendResponse(req.ID, ToolListResult{
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
			},
		})
	case "tools/call":
		handleToolCall(req)
	default:
		if req.ID != nil {
			sendError(req.ID, -32601, "Method not found", nil)
		}
	}
}

func handleToolCall(req JSONRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params", nil)
		return
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
	default:
		sendError(req.ID, -32602, "Unknown tool", nil)
		return
	}

	sendResponse(req.ID, result)
}

func sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}

func sendError(id interface{}, code int, message string, data interface{}) {
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
	fmt.Println(string(out))
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
