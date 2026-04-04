package toon_test

import (
	"bytes"
	"strings"
	"testing"

	"digital.vasic.toon/pkg/toon"
)

type testStruct struct {
	Name  string   `json:"name"`
	Value int      `json:"value"`
	Items []string `json:"items"`
}

func TestMarshalUnmarshal(t *testing.T) {
	original := testStruct{Name: "test", Value: 42, Items: []string{"a", "b"}}

	data, err := toon.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded testStruct
	if err := toon.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Name != "test" || decoded.Value != 42 || len(decoded.Items) != 2 {
		t.Errorf("round-trip failed: got %+v", decoded)
	}
}

func TestEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := toon.NewEncoder(&buf)

	if err := enc.Encode(testStruct{Name: "hello", Value: 1}); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("encoder produced no output")
	}
}

func TestDecoder(t *testing.T) {
	input := `{"name":"hello","value":1,"items":["x"]}`
	dec := toon.NewDecoder(strings.NewReader(input))

	var v testStruct
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if v.Name != "hello" || v.Value != 1 {
		t.Errorf("decoded wrong: %+v", v)
	}
}

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

func TestContentType(t *testing.T) {
	if toon.ContentType != "application/toon" {
		t.Errorf("ContentType = %q, want application/toon", toon.ContentType)
	}
}

func TestTokenEstimate(t *testing.T) {
	data := []byte("hello world test string")
	tokens := toon.TokenEstimate(data)
	if tokens <= 0 {
		t.Error("token estimate should be positive")
	}
}

func TestCompare(t *testing.T) {
	v := testStruct{Name: "test", Value: 42, Items: []string{"a", "b", "c"}}
	comp, err := toon.Compare(v)
	if err != nil {
		t.Fatalf("Compare error: %v", err)
	}
	if comp.JSONBytes <= 0 {
		t.Error("JSON bytes should be positive")
	}
	if comp.TOONBytes <= 0 {
		t.Error("TOON bytes should be positive")
	}
}

func TestMarshalIndent(t *testing.T) {
	v := testStruct{Name: "test", Value: 1}
	data, err := toon.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent error: %v", err)
	}
	if !bytes.Contains(data, []byte("\n")) {
		t.Error("indented output should contain newlines")
	}
}
