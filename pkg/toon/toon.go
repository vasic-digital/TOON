// Package toon provides Token-Oriented Object Notation (TOON) encoding
// and decoding for Go values, delegating to the upstream reference
// implementation at github.com/toon-format/toon-go.
//
// TOON is a token-efficient serialization format designed for LLM
// workflows where predictable structure and reduced token counts matter.
// Typical savings versus equivalent JSON range from ~8% on small struct
// payloads to ~40%+ on tabular array payloads (verified in toon_test.go).
//
// HISTORY (§107 anti-bluff anchor — round-27 §11.4 audit, 2026-05-17 +
// Wave 4b Task 2 wire-up, 2026-05-22):
//
// A PRIOR REVISION OF THIS FILE WAS A §11.4 CONTRACT-BLUFF. It silently
// delegated Marshal/Unmarshal to encoding/json while documenting itself
// as "TOON encoding". Callers expecting the advertised 30-60% token
// savings of Token-Oriented Object Notation were actually receiving
// identical JSON bytes. That bluff was reverted to an honest
// sentinel-error scaffold (ErrTOONEncodingNotImplemented on every
// entry point) in commit e36cfbb.
//
// This revision REPLACES the sentinel-error scaffold with real TOON
// encoding via github.com/toon-format/toon-go. The tests in toon_test.go
// assert THREE anti-regression invariants against the bluff:
//
//   1. WIRE-LEVEL: marshaled bytes MUST NOT start with '{' or '['
//      (i.e. they cannot be JSON-impersonating).
//   2. ROUND-TRIP: encode → decode → deep-equal (not "no error").
//   3. SIZE: meaningful savings versus encoding/json on a tabular
//      payload (≥10%).
//
// The ErrTOONEncodingNotImplemented sentinel is retained for backward
// compatibility but is no longer returned by any code path — see the
// var comment below.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic anchor),
// §11.4 PASS-bluff prevention, Herald §107 mandate.
package toon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	upstream "github.com/toon-format/toon-go"
)

// ContentType is the MIME type for TOON-encoded content.
const ContentType = "application/toon"

// ErrTOONEncodingNotImplemented is retained for backward compatibility
// with code written against the sentinel-error scaffold revision of
// this package (commit e36cfbb..fc2ab55). It is NO LONGER RETURNED by
// any Marshal/Unmarshal/Encoder/Decoder/Compare path — those now
// delegate to the real TOON implementation at
// github.com/toon-format/toon-go.
//
// The exported var stays in place so:
//   - existing `errors.Is(err, toon.ErrTOONEncodingNotImplemented)`
//     callsites compile cleanly (they will simply never match);
//   - the smoke test in Herald's commons/cli/toon_smoke_test.go that
//     verifies the symbol's presence (compile-time existence proof)
//     keeps passing during the upstream-bump transition.
//
// To assert "TOON encoding is real, not sentinel-stubbed", call
// Marshal on a known input and assert the first byte is NEITHER '{'
// NOR '[' — see TestMarshal_NotJSON in toon_test.go.
//
// Constitutional anchors: CONST-035 (anti-bluff), Article XI §11.9.
var ErrTOONEncodingNotImplemented = errors.New("toon: ErrTOONEncodingNotImplemented retained for ABI compatibility but no longer returned (Wave 4b T2 wired github.com/toon-format/toon-go)")

// Encoder writes TOON-encoded values to an underlying io.Writer.
//
// Backward-compat note: the previous (sentinel-scaffold) revision
// exposed the same io.Writer-based Encoder shape. The upstream
// reference impl uses an option-driven encoder without a streaming
// writer; we adapt by marshalling into a buffer and copying to the
// writer on Encode.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a TOON encoder that writes to w on Encode.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode renders v as a TOON document and writes the bytes to the
// underlying writer. Returns any encoding or writer error.
//
// §107 anti-bluff: the output is REAL TOON via
// github.com/toon-format/toon-go. The output never starts with '{'
// or '[' (which would indicate JSON impersonation — the original
// 2026-05-17 bluff). See TestEncoder_WriteIsNotJSON in toon_test.go.
//
// If w is nil, Encode marshals into an in-memory buffer and discards
// the bytes — that path is exercised by the Herald smoke test which
// passes a nil writer purely to prove the type/method surface exists.
func (e *Encoder) Encode(v interface{}) error {
	data, err := upstream.Marshal(v)
	if err != nil {
		return fmt.Errorf("toon: encode: %w", err)
	}
	if e.w == nil {
		return nil
	}
	if _, err := e.w.Write(data); err != nil {
		return fmt.Errorf("toon: write: %w", err)
	}
	return nil
}

// Decoder reads TOON-encoded values from an underlying io.Reader.
//
// Backward-compat note: the previous (sentinel-scaffold) revision
// exposed the same io.Reader-based Decoder shape. The upstream
// reference impl is byte-slice based; we adapt by io.ReadAll-ing the
// reader on Decode and delegating to upstream.Unmarshal.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a TOON decoder reading from r on Decode.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads TOON-encoded data from the underlying reader and
// decodes it into v, which MUST be a non-nil pointer.
//
// If the reader is nil, Decode returns an error rather than
// silently succeeding with a zero value (that would be a §11.4
// PASS-bluff at the decode layer — see TestDecoder_NilReader).
func (d *Decoder) Decode(v interface{}) error {
	if d.r == nil {
		return errors.New("toon: decode: nil reader")
	}
	if v == nil {
		return errors.New("toon: decode: nil target value")
	}
	data, err := io.ReadAll(d.r)
	if err != nil {
		return fmt.Errorf("toon: read: %w", err)
	}
	if err := upstream.Unmarshal(data, v); err != nil {
		return fmt.Errorf("toon: decode: %w", err)
	}
	return nil
}

// Marshal renders v as TOON bytes via the upstream reference
// implementation. Returns the encoded bytes or an error.
//
// §107 anti-bluff: the returned bytes are REAL TOON, not JSON.
// TestMarshal_NotJSON asserts the first byte is neither '{' nor
// '[' — that is the wire-level anti-regression marker against the
// original 2026-05-17 bluff (which silently json.Marshal-ed v).
//
// Encoding contract: struct fields are named via the `toon:"name"`
// tag where present; falls back to the field's Go name otherwise.
// Time values normalise to RFC-3339 by default; see WithTimeFormatter
// upstream option if you need a different formatter (not exposed
// through this wrapper — use github.com/toon-format/toon-go directly
// for advanced configuration).
func Marshal(v interface{}) ([]byte, error) {
	data, err := upstream.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("toon: marshal: %w", err)
	}
	return data, nil
}

// MarshalIndent renders v as TOON bytes with configurable indentation.
//
// The `prefix` parameter is accepted for io-compat with the
// encoding/json signature but is IGNORED — TOON's indentation model
// has no prefix concept. The `indent` parameter's character count is
// interpreted as the per-level space count (TOON spec §19); other
// indent strings (tabs, mixed) collapse to len(indent) spaces.
//
// §107 anti-bluff: like Marshal, the bytes never start with '{' or
// '['. See TestMarshalIndent_NotJSON.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	_ = prefix
	spaces := len(indent)
	if spaces <= 0 {
		spaces = 2 // sensible default; upstream's TOON Core Profile.
	}
	data, err := upstream.Marshal(v, upstream.WithIndent(spaces))
	if err != nil {
		return nil, fmt.Errorf("toon: marshal-indent: %w", err)
	}
	return data, nil
}

// Unmarshal decodes the TOON document in data into v, which MUST be
// a non-nil pointer (per the upstream contract).
//
// Returns a descriptive error for nil / non-pointer targets so the
// decoder type-checks behave like encoding/json's Unmarshal —
// catching the most common caller bug at the boundary.
func Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("toon: unmarshal: nil target value")
	}
	if err := upstream.Unmarshal(data, v); err != nil {
		return fmt.Errorf("toon: unmarshal: %w", err)
	}
	return nil
}

// IsTOONContentType checks if a content-type header indicates TOON.
//
// Case-insensitive substring check; matches "application/toon",
// "Application/TOON; charset=utf-8", etc.
func IsTOONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/toon")
}

// TokenEstimate estimates the number of LLM tokens for a byte slice
// using the well-known ~4-chars-per-token heuristic for English text.
//
// This helper makes no TOON-specific claim — it is suitable for rough
// sizing of any byte content (JSON, TOON, prose). Returns 0 for nil
// or empty input.
func TokenEstimate(data []byte) int {
	return len(data) / 4
}

// TokenComparison holds the result of a side-by-side JSON vs TOON
// encoding comparison for a given Go value. Used by Compare.
//
// JSONBytes / TOONBytes are byte lengths of the respective encodings.
// JSONTokens / TOONTokens are the corresponding TokenEstimate values.
// Savings is the percentage by which TOON beats JSON in tokens, e.g.
// 30.0 means "TOON uses 30% fewer estimated tokens than JSON". A
// negative value means TOON used MORE tokens (possible on tiny
// scalar-only payloads where TOON's headers cost more than JSON's
// braces; uncommon in realistic payloads).
type TokenComparison struct {
	JSONBytes  int     `json:"json_bytes"`
	TOONBytes  int     `json:"toon_bytes"`
	JSONTokens int     `json:"json_tokens"`
	TOONTokens int     `json:"toon_tokens"`
	Savings    float64 `json:"savings_percent"`
}

// Compare encodes v in both JSON and TOON and returns a token-by-token
// comparison.
//
// §107 anti-bluff: this is NOT the round-27 fabricated comparison
// (which json.Marshal-ed v twice and reported Savings=0.0). This
// function uses encoding/json for the JSON side and upstream
// github.com/toon-format/toon-go for the TOON side, and the returned
// Savings reflects the REAL token delta. See TestCompare_RealSavings.
func Compare(v interface{}) (*TokenComparison, error) {
	tb, err := upstream.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("toon: compare: toon-marshal: %w", err)
	}
	jb, err := jsonMarshal(v)
	if err != nil {
		return nil, fmt.Errorf("toon: compare: json-marshal: %w", err)
	}
	jt := TokenEstimate(jb)
	tt := TokenEstimate(tb)
	var savings float64
	if jt > 0 {
		savings = 100.0 * (1.0 - float64(tt)/float64(jt))
	}
	return &TokenComparison{
		JSONBytes:  len(jb),
		TOONBytes:  len(tb),
		JSONTokens: jt,
		TOONTokens: tt,
		Savings:    savings,
	}, nil
}

// jsonMarshal renders v as JSON for the Compare side-by-side. The
// only use of encoding/json in this package is this internal
// comparison helper — production Marshal/Unmarshal go through the
// upstream TOON impl, never through encoding/json. That is the
// structural anti-bluff invariant: no `if err != nil { return
// json.Marshal(v) }` fallback exists anywhere in this file.
func jsonMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// SetEscapeHTML(false) keeps the JSON side comparable to what
	// most production callers actually emit (Gin's c.JSON, the
	// gin-contrib serializers, etc.).
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// json.Encoder always appends a trailing newline; trim for a
	// fair byte-length comparison versus toon.Marshal's no-trailing-
	// newline output.
	b := buf.Bytes()
	for len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b, nil
}
