package rest

import (
	"encoding/json"
	"net/http"
)

// Meta contains pagination and metadata for the response.
type Meta struct {
	Limit int64   `json:"limit"`
	Skip  *int64  `json:"skip,omitempty"`
	Total *int64  `json:"total,omitempty"`
	Next  *string `json:"next,omitempty"`
	Prev  *string `json:"prev,omitempty"`
}

// Envelope is the standard response structure.
type Envelope struct {
	Data  any    `json:"data"`
	Meta  *Meta  `json:"meta,omitempty"`
	Error *Error `json:"error,omitempty"`
}

// Error represents an API error.
type Error struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// DefaultEncoder implements Encoder using encoding/json.
type DefaultEncoder struct{}

func (e DefaultEncoder) Write(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func writeJSON(w http.ResponseWriter, encoder Encoder, status int, data any, meta *Meta) {
	if encoder == nil {
		encoder = DefaultEncoder{}
	}
	w.WriteHeader(status)
	_ = encoder.Write(w, Envelope{
		Data: data,
		Meta: meta,
	})
}

func writeError(w http.ResponseWriter, encoder Encoder, status int, msg string, code string) {
	if encoder == nil {
		encoder = DefaultEncoder{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = encoder.Write(w, Envelope{
		Error: &Error{
			Message: msg,
			Code:    code,
		},
	})
}
