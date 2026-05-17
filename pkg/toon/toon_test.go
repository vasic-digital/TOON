package toon_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"digital.vasic.toon/pkg/toon"
)

type testStruct struct {
	Name  string   `json:"name"`
	Value int      `json:"value"`
	Items []string `json:"items"`
}

// TestMarshal_ReturnsSentinel asserts the round-27 §11.4 audit fix:
// Marshal() no longer silently delegates to json.Marshal; it returns
// the ErrTOONEncodingNotImplemented sentinel so callers cannot mistake
// JSON output for the advertised TOON token-savings encoding.
func TestMarshal_ReturnsSentinel(t *testing.T) {
	data, err := toon.Marshal(testStruct{Name: "test", Value: 42})
	if err == nil {
		t.Fatalf("Marshal: expected ErrTOONEncodingNotImplemented, got nil error and data=%q", data)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("Marshal: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
	if data != nil {
		t.Errorf("Marshal: expected nil data, got %q", data)
	}
}

// TestMarshalIndent_ReturnsSentinel mirrors TestMarshal but exercises
// the indented entry point that previously claimed "TOON-style minimal
// whitespace" while emitting json.MarshalIndent output.
func TestMarshalIndent_ReturnsSentinel(t *testing.T) {
	data, err := toon.MarshalIndent(testStruct{Name: "test", Value: 1}, "", "  ")
	if err == nil {
		t.Fatalf("MarshalIndent: expected ErrTOONEncodingNotImplemented, got nil error and data=%q", data)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("MarshalIndent: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
}

// TestUnmarshal_ReturnsSentinel asserts the decoder side of the round-27
// §11.4 audit fix.
func TestUnmarshal_ReturnsSentinel(t *testing.T) {
	var v testStruct
	err := toon.Unmarshal([]byte(`{"name":"test"}`), &v)
	if err == nil {
		t.Fatalf("Unmarshal: expected ErrTOONEncodingNotImplemented, got nil error and v=%+v", v)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("Unmarshal: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
}

// TestEncoder_ReturnsSentinel asserts Encoder.Encode surfaces the sentinel
// rather than silently writing JSON bytes to the underlying writer.
func TestEncoder_ReturnsSentinel(t *testing.T) {
	var buf bytes.Buffer
	enc := toon.NewEncoder(&buf)

	err := enc.Encode(testStruct{Name: "hello", Value: 1})
	if err == nil {
		t.Fatalf("Encoder.Encode: expected ErrTOONEncodingNotImplemented, got nil error and buf=%q", buf.String())
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("Encoder.Encode: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
	if buf.Len() != 0 {
		t.Errorf("Encoder.Encode: expected empty buffer on sentinel return, got %q", buf.String())
	}
}

// TestDecoder_ReturnsSentinel asserts Decoder.Decode surfaces the sentinel
// rather than silently json.Unmarshal-ing the input.
func TestDecoder_ReturnsSentinel(t *testing.T) {
	input := `{"name":"hello","value":1,"items":["x"]}`
	dec := toon.NewDecoder(strings.NewReader(input))

	var v testStruct
	err := dec.Decode(&v)
	if err == nil {
		t.Fatalf("Decoder.Decode: expected ErrTOONEncodingNotImplemented, got nil error and v=%+v", v)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("Decoder.Decode: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
}

// TestCompare_ReturnsSentinel asserts Compare no longer fabricates a
// Savings: 0.0 comparison by calling json.Marshal twice and pretending
// it had measured TOON vs JSON.
func TestCompare_ReturnsSentinel(t *testing.T) {
	v := testStruct{Name: "test", Value: 42, Items: []string{"a", "b", "c"}}
	comp, err := toon.Compare(v)
	if err == nil {
		t.Fatalf("Compare: expected ErrTOONEncodingNotImplemented, got nil error and comp=%+v", comp)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Errorf("Compare: error = %v, want errors.Is(_, ErrTOONEncodingNotImplemented)", err)
	}
	if comp != nil {
		t.Errorf("Compare: expected nil comparison on sentinel return, got %+v", comp)
	}
}

// TestIsTOONContentType — header inspection helper, not part of the
// encoding-bluff scope; still safe and meaningful.
func TestIsTOONContentType(t *testing.T) {
	if !toon.IsTOONContentType("application/toon") {
		t.Error("should match application/toon")
	}
	if !toon.IsTOONContentType("Application/TOON; charset=utf-8") {
		t.Error("should match case-insensitive")
	}
	if toon.IsTOONContentType("application/json") {
		t.Error("should not match application/json")
	}
}

// TestContentType — constant exposure, no bluff.
func TestContentType(t *testing.T) {
	if toon.ContentType != "application/toon" {
		t.Errorf("ContentType = %q, want application/toon", toon.ContentType)
	}
}

// TestTokenEstimate — heuristic char/4 helper that makes no
// TOON-specific claim; safe and meaningful.
func TestTokenEstimate(t *testing.T) {
	data := []byte("hello world test string")
	tokens := toon.TokenEstimate(data)
	if tokens <= 0 {
		t.Error("token estimate should be positive")
	}
}
