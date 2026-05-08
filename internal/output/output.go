package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

type Error struct {
	Code     string            `json:"code"`
	Category string            `json:"category,omitempty"`
	Message  string            `json:"message"`
	Hint     string            `json:"hint,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

type Meta struct {
	Command    string `json:"command"`
	TS         string `json:"ts"`
	DurationMS int64  `json:"duration_ms"`
}

type Envelope struct {
	OK      bool   `json:"ok"`
	Version string `json:"version"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    Meta   `json:"meta"`
}

const ContractVersion = "1"

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func WriteSuccess(w io.Writer, command string, startedAt time.Time, value any) error {
	return WriteJSON(w, Envelope{
		OK:      true,
		Version: ContractVersion,
		Data:    value,
		Meta:    buildMeta(command, startedAt),
	})
}

func WriteErrorEnvelope(w io.Writer, command string, startedAt time.Time, err Error) error {
	return WriteJSON(w, Envelope{
		OK:      false,
		Version: ContractVersion,
		Error:   &err,
		Meta:    buildMeta(command, startedAt),
	})
}

func WriteError(w io.Writer, format Format, err Error) error {
	if format == FormatJSON {
		return WriteErrorEnvelope(w, "", time.Time{}, err)
	}
	_, writeErr := fmt.Fprintf(w, "Error: %s\n", err.Message)
	return writeErr
}

func buildMeta(command string, startedAt time.Time) Meta {
	now := time.Now().UTC()
	durationMS := int64(0)
	if !startedAt.IsZero() {
		durationMS = time.Since(startedAt).Milliseconds()
		if durationMS < 0 {
			durationMS = 0
		}
	}
	return Meta{
		Command:    command,
		TS:         now.Format(time.RFC3339Nano),
		DurationMS: durationMS,
	}
}
