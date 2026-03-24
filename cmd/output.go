package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

type commandStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func writeCommandError(w io.Writer, message string) {
	writeJSON(w, commandStatus{Status: "error", Message: message})
}

func writeJSON(w io.Writer, value interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(w, "Error encoding JSON output: %v\n", err)
	}
}
