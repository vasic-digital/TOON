package toon

import (
	"fmt"
	"io"
	"strings"
)

// ContentType is the MIME type for TOON-encoded content.
const ContentType = "application/toon"

// ErrTOONEncodingNotImplemented is returned by every public Marshal /
// Unmarshal / MarshalIndent / Compare entry point in this module. The
// previous implementation of those functions silently delegated to
// encoding/json while documenting themselves as "TOON encoding" —
// a CRITICAL contract-bluff at the module-purpose layer (round-27
// §11.4 audit, 2026-05-17): every caller believed it was getting
// the advertised 30-60% token savings of Token-Oriented Object
// Notation but was actually receiving identical JSON bytes.
//
// Until native TOON encoding is wired (via upstream toon-format/toon-go
// or a hand-written encoder), the module's public API surfaces this
// sentinel rather than fabricate success. Callers that need a JSON
// encoding right now MUST use encoding/json directly — there is no
// dishonest convenience wrapper here.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic anchor).
var ErrTOONEncodingNotImplemented = fmt.Errorf("toon: native TOON encoding is not yet implemented in this module; the previous Marshal()/Unmarshal()/MarshalIndent() silently fell back to JSON despite doc-comments claiming TOON encoding (§11.4 PASS-bluff — Article XI §11.9 forensic-anchor pattern). Use encoding/json directly or wait for toon-format/toon-go upstream integration")

// Encoder writes TOON-encoded values.
//
// HISTORICAL BLUFF (resolved round-27, 2026-05-17): previous revisions
// claimed "Falls back to a token-efficient JSON representation when a
// native TOON encoder isn't available" — that fallback WAS the bluff.
// Encode now returns ErrTOONEncodingNotImplemented. See package-level
// docstring on ErrTOONEncodingNotImplemented for details.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a TOON encoder writing to w. Encode() will return
// ErrTOONEncodingNotImplemented until native TOON encoding is wired.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the TOON encoding of v.
//
// Returns ErrTOONEncodingNotImplemented unconditionally — native TOON
// encoding is not yet implemented in this module. See the package-level
// ErrTOONEncodingNotImplemented documentation for the round-27 §11.4
// audit context.
func (e *Encoder) Encode(v interface{}) error {
	_ = v
	return ErrTOONEncodingNotImplemented
}

// Decoder reads TOON-encoded values.
//
// HISTORICAL BLUFF (resolved round-27, 2026-05-17): previous revisions
// "accepted both TOON and JSON input" by silently delegating to
// encoding/json — the TOON acceptance was fabricated. Decode now
// returns ErrTOONEncodingNotImplemented.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a TOON decoder reading from r. Decode() will
// return ErrTOONEncodingNotImplemented until native TOON decoding is
// wired.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads TOON-encoded data into v.
//
// Returns ErrTOONEncodingNotImplemented unconditionally — see the
// package-level sentinel documentation.
func (d *Decoder) Decode(v interface{}) error {
	_ = v
	return ErrTOONEncodingNotImplemented
}

// Marshal encodes v as TOON bytes.
//
// Returns ErrTOONEncodingNotImplemented unconditionally. The previous
// implementation called json.Marshal and lied about it in the doc
// comment — round-27 §11.4 audit removed that bluff. Use encoding/json
// directly if you need JSON; wait for upstream TOON encoder
// integration if you need actual TOON.
func Marshal(v interface{}) ([]byte, error) {
	_ = v
	return nil, ErrTOONEncodingNotImplemented
}

// MarshalIndent encodes v as human-readable TOON.
//
// Returns ErrTOONEncodingNotImplemented unconditionally. The previous
// implementation called json.MarshalIndent and claimed "TOON-style
// minimal whitespace" — that was a §11.4 PASS-bluff (the JSON
// indentation is not TOON's minimal-whitespace style at all).
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	_ = v
	_ = prefix
	_ = indent
	return nil, ErrTOONEncodingNotImplemented
}

// Unmarshal decodes TOON bytes into v.
//
// Returns ErrTOONEncodingNotImplemented unconditionally. The previous
// implementation called json.Unmarshal and claimed acceptance of
// "both TOON and JSON input" — round-27 §11.4 audit removed that
// bluff. Use encoding/json directly for JSON input.
func Unmarshal(data []byte, v interface{}) error {
	_ = data
	_ = v
	return ErrTOONEncodingNotImplemented
}

// IsTOONContentType checks if a content type header indicates TOON format.
//
// This helper is safe — it only inspects header strings and does not
// claim any encoding/decoding work.
func IsTOONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/toon")
}

// TokenEstimate estimates the number of LLM tokens for the given data.
//
// This is a heuristic helper (~4 chars per token for English text)
// and makes no claim about TOON vs JSON ratios. Suitable for rough
// sizing of arbitrary byte content.
func TokenEstimate(data []byte) int {
	// Rough estimate: ~4 chars per token for English text
	return len(data) / 4
}

// TokenComparison compares token counts between JSON and TOON encodings.
//
// HISTORICAL BLUFF (resolved round-27, 2026-05-17): previous revisions
// populated this struct by calling json.Marshal twice (once labelled
// "json", once labelled "toon"), then computing Savings = 0.0
// deterministically. The result was a structure whose every field
// pretended to be a meaningful comparison but was actually a no-op.
// Compare() now returns ErrTOONEncodingNotImplemented instead of
// fabricating comparison data.
type TokenComparison struct {
	JSONBytes  int     `json:"json_bytes"`
	TOONBytes  int     `json:"toon_bytes"`
	JSONTokens int     `json:"json_tokens"`
	TOONTokens int     `json:"toon_tokens"`
	Savings    float64 `json:"savings_percent"`
}

// Compare encodes v in both formats and reports the difference.
//
// Returns (nil, ErrTOONEncodingNotImplemented) unconditionally — the
// previous implementation marshalled v with json.Marshal twice and
// reported Savings: 0.0 as if it had performed a meaningful TOON-vs-
// JSON comparison. Round-27 §11.4 audit replaced that fabricated
// success with the honest sentinel.
func Compare(v interface{}) (*TokenComparison, error) {
	_ = v
	return nil, ErrTOONEncodingNotImplemented
}
