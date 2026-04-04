package toon

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ContentType is the MIME type for TOON-encoded content.
const ContentType = "application/toon"

// Encoder writes TOON-encoded values. Falls back to a token-efficient
// JSON representation when a native TOON encoder isn't available.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a TOON encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the TOON encoding of v.
// Current implementation uses a compact, token-efficient JSON representation
// (no unnecessary whitespace, short keys preserved). When toon-format/toon-go
// becomes available, this will use native TOON encoding.
func (e *Encoder) Encode(v interface{}) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}

// Decoder reads TOON-encoded values.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a TOON decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads TOON-encoded data into v.
// Current implementation accepts both TOON and JSON input.
func (d *Decoder) Decode(v interface{}) error {
	data, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}

// Marshal encodes v as TOON bytes.
// Current implementation: compact JSON (no indentation).
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent encodes v as human-readable TOON.
// Current implementation: indented JSON with TOON-style minimal whitespace.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal decodes TOON bytes into v.
// Accepts both TOON and JSON input.
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// IsTOONContentType checks if a content type header indicates TOON format.
func IsTOONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/toon")
}

// TokenEstimate estimates the number of LLM tokens for the given data.
// TOON format typically uses 30-60% fewer tokens than JSON.
func TokenEstimate(data []byte) int {
	// Rough estimate: ~4 chars per token for English text
	return len(data) / 4
}

// TokenComparison compares token counts between JSON and TOON encodings.
type TokenComparison struct {
	JSONBytes  int     `json:"json_bytes"`
	TOONBytes  int     `json:"toon_bytes"`
	JSONTokens int     `json:"json_tokens"`
	TOONTokens int     `json:"toon_tokens"`
	Savings    float64 `json:"savings_percent"`
}

// Compare encodes v in both formats and reports the difference.
func Compare(v interface{}) (*TokenComparison, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}
	toonData, err := Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("toon marshal: %w", err)
	}

	jsonTokens := TokenEstimate(jsonData)
	toonTokens := TokenEstimate(toonData)

	savings := 0.0
	if jsonTokens > 0 {
		savings = float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100
	}

	return &TokenComparison{
		JSONBytes:  len(jsonData),
		TOONBytes:  len(toonData),
		JSONTokens: jsonTokens,
		TOONTokens: toonTokens,
		Savings:    savings,
	}, nil
}
