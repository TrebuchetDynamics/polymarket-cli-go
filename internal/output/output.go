package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error Error `json:"error"`
}

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func WriteError(w io.Writer, format Format, err Error) error {
	if format == FormatJSON {
		return WriteJSON(w, errorEnvelope{Error: err})
	}
	_, writeErr := fmt.Fprintf(w, "Error: %s\n", err.Message)
	return writeErr
}
