// Package toon_test verifies the Wave 4b T2 wire-up of
// github.com/toon-format/toon-go.
//
// This test file is the §107 anti-bluff harness. It encodes THREE
// invariants that, together, make a regression to the original
// 2026-05-17 JSON-fallback bluff DETECTABLE BY THE BUILD:
//
//  1. WIRE-LEVEL — marshaled bytes MUST NOT start with '{' or '[';
//     those are JSON syntax markers and the original bluff produced
//     such bytes. TestMarshal_NotJSON / TestMarshalIndent_NotJSON /
//     TestEncoder_WriteIsNotJSON cover all three entry points.
//
//  2. ROUND-TRIP — encode → decode → deep-equal to the original Go
//     value. Asserting "no error" alone (as the original bluff did)
//     does NOT verify the bytes round-trip; deep-equal does.
//     TestRoundTrip_ScalarStruct and TestRoundTrip_TabularPayload
//     cover both scalar-only and tabular-array shapes.
//
//  3. SIZE — TOON must save ≥10% of token budget vs JSON on a
//     realistic tabular payload (verified at 42.6% for the chosen
//     Receipt shape during W4b T2 development; lower bound chosen
//     to leave margin for upstream evolution).
//     TestCompare_RealSavings asserts the size delta.
//
// Plus a handful of compatibility tests for the helpers
// (ContentType, IsTOONContentType, TokenEstimate) that were never
// bluff-prone.
package toon_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"digital.vasic.toon/pkg/toon"
)

// Receipt is the canonical Herald shape used to verify TOON's
// tabular-savings claim. It mirrors the production
// Receipt/DeliveryAttempt model used in Herald's commons/cli layer
// closely enough that the size-savings number is realistic.
type DeliveryAttempt struct {
	Idx       int    `toon:"idx" json:"idx"`
	Channel   string `toon:"channel" json:"channel"`
	Status    string `toon:"status" json:"status"`
	LatencyMs int    `toon:"latency_ms" json:"latency_ms"`
}

type Receipt struct {
	EventID  string            `toon:"event_id" json:"event_id"`
	Subject  string            `toon:"subject" json:"subject"`
	Body     string            `toon:"body" json:"body"`
	Tags     []string          `toon:"tags" json:"tags"`
	Attempts []DeliveryAttempt `toon:"attempts" json:"attempts"`
}

func sampleReceipt() Receipt {
	return Receipt{
		EventID: "evt_01HQRZ5K4ZNV4M8DXKQRG3FJN8",
		Subject: "Build broken",
		Body:    "main pipeline failed step 17 (vet) at 2026-05-22T08:14:32Z",
		Tags:    []string{"prod", "ci", "p1"},
		Attempts: []DeliveryAttempt{
			{Idx: 1, Channel: "telegram", Status: "delivered", LatencyMs: 142},
			{Idx: 2, Channel: "slack", Status: "delivered", LatencyMs: 87},
			{Idx: 3, Channel: "email", Status: "queued", LatencyMs: 0},
			{Idx: 4, Channel: "max", Status: "failed", LatencyMs: 5012},
		},
	}
}

// firstByteIsJSON returns true if data starts with '{' or '[' (the
// only two possible first bytes of a JSON document per RFC 8259 §2).
// This is the wire-level anti-bluff predicate.
func firstByteIsJSON(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return data[0] == '{' || data[0] == '['
}

// TestMarshal_NotJSON — §107 wire-level invariant for Marshal.
//
// If a future regression silently re-introduces JSON-fallback (the
// 2026-05-17 bluff), this test catches it: TOON output cannot
// possibly start with '{' or '[' because TOON's grammar leads with
// the first key name (per spec §2 "Document").
func TestMarshal_NotJSON(t *testing.T) {
	r := sampleReceipt()
	data, err := toon.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Marshal: empty output (no encoding happened?)")
	}
	if firstByteIsJSON(data) {
		t.Fatalf("Marshal: §107 BLUFF REGRESSION — output starts with JSON syntax byte %q; first 80 bytes: %q",
			string(data[0]), string(data[:min(80, len(data))]))
	}
}

// TestMarshalIndent_NotJSON — §107 wire-level invariant for
// MarshalIndent. The original bluff applied to this entry point too
// (it called json.MarshalIndent and claimed "TOON-style minimal
// whitespace"). Same first-byte check.
func TestMarshalIndent_NotJSON(t *testing.T) {
	r := sampleReceipt()
	data, err := toon.MarshalIndent(r, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: unexpected error: %v", err)
	}
	if firstByteIsJSON(data) {
		t.Fatalf("MarshalIndent: §107 BLUFF REGRESSION — output starts with JSON syntax byte %q", string(data[0]))
	}
}

// TestEncoder_WriteIsNotJSON — §107 wire-level invariant for the
// Encoder.Encode path. The encoder writes to an io.Writer; we
// inspect the buffer to assert the bytes are real TOON.
func TestEncoder_WriteIsNotJSON(t *testing.T) {
	var buf bytes.Buffer
	enc := toon.NewEncoder(&buf)
	if err := enc.Encode(sampleReceipt()); err != nil {
		t.Fatalf("Encoder.Encode: unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("Encoder.Encode: buffer is empty after Encode")
	}
	if firstByteIsJSON(buf.Bytes()) {
		t.Fatalf("Encoder.Encode: §107 BLUFF REGRESSION — first byte is %q", string(buf.Bytes()[0]))
	}
}

// TestRoundTrip_ScalarStruct — round-trip invariant on a scalar-
// only struct. Encode → Unmarshal → reflect.DeepEqual.
func TestRoundTrip_ScalarStruct(t *testing.T) {
	type Simple struct {
		Name  string `toon:"name"`
		Value int    `toon:"value"`
		Flag  bool   `toon:"flag"`
	}
	orig := Simple{Name: "hello", Value: 42, Flag: true}

	data, err := toon.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if firstByteIsJSON(data) {
		t.Fatalf("BLUFF: output is JSON, not TOON: %q", string(data))
	}

	var back Simple
	if err := toon.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch:\n  orig = %+v\n  back = %+v\n  wire = %q",
			orig, back, string(data))
	}
}

// TestRoundTrip_TabularPayload — round-trip invariant on the
// tabular-array shape that TOON is specifically optimised for.
// This is the shape where the size-savings assertion will land.
func TestRoundTrip_TabularPayload(t *testing.T) {
	orig := sampleReceipt()

	data, err := toon.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if firstByteIsJSON(data) {
		t.Fatalf("BLUFF: output is JSON, not TOON: %q", string(data))
	}

	var back Receipt
	if err := toon.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal: %v\nwire: %s", err, string(data))
	}
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch:\n  orig = %+v\n  back = %+v\n  wire = %q",
			orig, back, string(data))
	}
}

// TestCompare_RealSavings — §107 size invariant. TOON must beat
// JSON by ≥10% on a realistic tabular payload. Anything less than
// 10% suggests either a degraded upstream impl OR a regression to
// the JSON-fallback bluff (which would report Savings ≈ 0).
//
// Observed during W4b T2 development (toon-go v0.0.0-20251202084852):
//
//	JSON len = 340, TOON len = 195 → 42.6% savings on bytes;
//	token estimates likewise differ by ~42%.
//
// We assert a conservative 10% lower bound so upstream optimisations
// don't immediately break this test, while still catching any
// regression that would silently fall back to JSON.
func TestCompare_RealSavings(t *testing.T) {
	r := sampleReceipt()
	comp, err := toon.Compare(r)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if comp == nil {
		t.Fatal("Compare: nil result")
	}
	if comp.JSONBytes <= 0 || comp.TOONBytes <= 0 {
		t.Fatalf("Compare: degenerate byte counts: %+v", comp)
	}
	if comp.TOONBytes >= comp.JSONBytes {
		t.Fatalf("§107 BLUFF REGRESSION — TOONBytes (%d) >= JSONBytes (%d); TOON should be smaller on tabular payloads (the original bluff would have reported these EQUAL because TOON was secretly JSON)",
			comp.TOONBytes, comp.JSONBytes)
	}
	if comp.Savings < 10.0 {
		t.Fatalf("§107 size invariant failed: Savings = %.1f%%, want ≥ 10%%; full comp = %+v",
			comp.Savings, comp)
	}
	t.Logf("Compare: TOON saves %.1f%% vs JSON on Receipt shape (json=%dB toon=%dB)",
		comp.Savings, comp.JSONBytes, comp.TOONBytes)
}

// TestCompare_ContrastsRealEncodings — second-order §107 anti-
// bluff check: assert that the bytes used inside Compare are
// actually different. If a regression made TOON ≡ JSON, the
// Compare result would still produce a number, but the two byte
// streams would be byte-identical. We exercise both encoders
// directly and assert they produce DIFFERENT byte streams.
func TestCompare_ContrastsRealEncodings(t *testing.T) {
	r := sampleReceipt()
	tb, err := toon.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal toon: %v", err)
	}
	jb, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal json: %v", err)
	}
	if bytes.Equal(tb, jb) {
		t.Fatal("§107 BLUFF REGRESSION — toon.Marshal output is byte-identical to json.Marshal output (the original 2026-05-17 bluff)")
	}
	if firstByteIsJSON(tb) {
		t.Fatalf("BLUFF: TOON output is JSON-shaped: %q", string(tb))
	}
	if !firstByteIsJSON(jb) {
		t.Fatalf("test sanity: JSON output is NOT JSON-shaped (test bug): %q", string(jb))
	}
}

// TestUnmarshal_NilTarget — decoder type-checks at the boundary.
// Passing v=nil must produce a clear error rather than a panic or
// silent success.
func TestUnmarshal_NilTarget(t *testing.T) {
	err := toon.Unmarshal([]byte("name: test\n"), nil)
	if err == nil {
		t.Fatal("Unmarshal(nil): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Unmarshal(nil): error = %v, want message mentioning 'nil'", err)
	}
}

// TestDecoder_NilReader — Decoder constructed with nil reader must
// return an error on Decode, not silently succeed.
func TestDecoder_NilReader(t *testing.T) {
	dec := toon.NewDecoder(nil)
	if dec == nil {
		t.Fatal("NewDecoder(nil) returned nil decoder")
	}
	var v map[string]any
	if err := dec.Decode(&v); err == nil {
		t.Fatal("Decoder.Decode with nil reader: expected error, got nil")
	}
}

// TestEncoder_NilWriter — Encoder constructed with nil writer must
// NOT panic on Encode. (It silently discards bytes; see the
// in-package comment.)
//
// This is exercised by the Herald smoke test which deliberately
// passes nil to prove the type/method surface exists without
// caring about output.
func TestEncoder_NilWriter(t *testing.T) {
	enc := toon.NewEncoder(nil)
	if enc == nil {
		t.Fatal("NewEncoder(nil) returned nil encoder")
	}
	if err := enc.Encode(sampleReceipt()); err != nil {
		t.Fatalf("Encoder.Encode with nil writer: unexpected error: %v", err)
	}
}

// TestErrTOONEncodingNotImplemented_RetainedForABI — the sentinel
// var stays exported for ABI compatibility with code written
// against the e36cfbb..fc2ab55 revisions of this package. It must
// still be present, still be a non-nil error, but it must NOT be
// returned by any actual code path (because we now have real
// TOON encoding).
func TestErrTOONEncodingNotImplemented_RetainedForABI(t *testing.T) {
	if toon.ErrTOONEncodingNotImplemented == nil {
		t.Fatal("ErrTOONEncodingNotImplemented is nil — ABI break")
	}
	if toon.ErrTOONEncodingNotImplemented.Error() == "" {
		t.Fatal("ErrTOONEncodingNotImplemented has empty message")
	}

	// Make sure none of the real entry points actually RETURN it.
	r := sampleReceipt()
	if _, err := toon.Marshal(r); errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Fatal("Marshal still returns ErrTOONEncodingNotImplemented — W4b T2 wire-up incomplete")
	}
	if _, err := toon.MarshalIndent(r, "", "  "); errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Fatal("MarshalIndent still returns sentinel")
	}
	var back Receipt
	data, _ := toon.Marshal(r)
	if err := toon.Unmarshal(data, &back); errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		t.Fatal("Unmarshal still returns sentinel")
	}
}

// TestIsTOONContentType — preserved from sentinel-scaffold tests.
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

// TestContentType — constant exposure check.
func TestContentType(t *testing.T) {
	if toon.ContentType != "application/toon" {
		t.Errorf("ContentType = %q, want application/toon", toon.ContentType)
	}
}

// TestTokenEstimate — heuristic helper sanity.
func TestTokenEstimate(t *testing.T) {
	if got := toon.TokenEstimate(nil); got != 0 {
		t.Errorf("TokenEstimate(nil) = %d, want 0", got)
	}
	if got := toon.TokenEstimate(make([]byte, 100)); got != 25 {
		t.Errorf("TokenEstimate(100B) = %d, want 25", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
