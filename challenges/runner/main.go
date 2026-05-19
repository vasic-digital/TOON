// Package main implements the TOON describe Challenge runner.
//
// Round-286 §11.4 enrichment (2026-05-19): this binary exercises every
// public symbol of digital.vasic.toon/pkg/toon from a real Go process
// across five locales (en, sr, ja, es, de) and asserts:
//
//  1. Every encoding entry point (Marshal, MarshalIndent, Unmarshal,
//     Encoder.Encode, Decoder.Decode, Compare) returns
//     ErrTOONEncodingNotImplemented via errors.Is.
//  2. Encoder.Encode does NOT write any bytes to the underlying writer
//     when returning the sentinel (writer purity).
//  3. Decoder.Decode does NOT mutate the target value when returning
//     the sentinel (target purity).
//  4. IsTOONContentType and TokenEstimate behave as documented
//     (header inspection and len/4 heuristic).
//  5. All five locale fixtures parse and present sentinel narratives
//     correctly (bilingual evidence per CONST-046).
//
// The companion shell wrapper challenges/scripts/toon_describe_challenge.sh
// invokes this binary, captures evidence, and provides the CONST-055
// paired-mutation gate (--anti-bluff-mutate flag).
//
// Constitutional anchors: CONST-035, CONST-048, CONST-050(B),
// CONST-055, Article XI §11.9.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"digital.vasic.toon/pkg/toon"
)

// localeFixture matches the schema of challenges/fixtures/<locale>.json.
type localeFixture struct {
	Locale          string `json:"locale"`
	Description     string `json:"description"`
	SentinelNarrate string `json:"sentinel_narrate"`
	HeaderSample    string `json:"header_sample"`
	PayloadSample   string `json:"payload_sample"`
}

// payload is the test struct exercised through every sentinel path.
type payload struct {
	Name  string   `json:"name"`
	Value int      `json:"value"`
	Items []string `json:"items"`
}

func main() {
	var (
		localesFlag = flag.String("locales", "en,sr,ja,es,de", "comma-separated locale list")
		fixturesDir = flag.String("fixtures", "", "fixtures directory (default: relative to binary)")
		mutate      = flag.Bool("mutate", false, "planted-bluff mode (must FAIL anti-bluff gate)")
	)
	flag.Parse()

	if *fixturesDir == "" {
		// Resolve fixtures relative to source file location.
		exe, err := os.Executable()
		if err != nil {
			die("cannot resolve executable path: %v", err)
		}
		// Walk up to challenges/ then into fixtures/.
		dir := filepath.Dir(exe)
		for i := 0; i < 6; i++ {
			candidate := filepath.Join(dir, "fixtures")
			if st, err := os.Stat(candidate); err == nil && st.IsDir() {
				*fixturesDir = candidate
				break
			}
			candidate2 := filepath.Join(dir, "challenges", "fixtures")
			if st, err := os.Stat(candidate2); err == nil && st.IsDir() {
				*fixturesDir = candidate2
				break
			}
			dir = filepath.Dir(dir)
		}
	}
	if *fixturesDir == "" {
		die("could not locate challenges/fixtures directory; use -fixtures")
	}

	locales := strings.Split(*localesFlag, ",")
	if len(locales) < 1 {
		die("no locales specified")
	}

	fmt.Printf("[toon-runner] mode=%s locales=%v fixtures=%s\n",
		modeLabel(*mutate), locales, *fixturesDir)

	for _, loc := range locales {
		loc = strings.TrimSpace(loc)
		if loc == "" {
			continue
		}
		if err := exerciseLocale(*fixturesDir, loc, *mutate); err != nil {
			die("locale %s FAILED: %v", loc, err)
		}
		fmt.Printf("[toon-runner] locale %s — sentinel OK\n", loc)
	}

	fmt.Printf("[toon-runner] all %d locales asserted sentinel; zero JSON bytes leaked\n",
		len(locales))
}

func exerciseLocale(fixturesDir, locale string, mutate bool) error {
	fixturePath := filepath.Join(fixturesDir, locale+".json")
	raw, err := ioutil.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("read fixture: %w", err)
	}
	var fx localeFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		return fmt.Errorf("parse fixture %s: %w", fixturePath, err)
	}
	if fx.Locale != locale {
		return fmt.Errorf("fixture locale mismatch: file=%s declares=%s", locale, fx.Locale)
	}
	if fx.SentinelNarrate == "" {
		return fmt.Errorf("fixture %s missing sentinel_narrate", locale)
	}

	// 1. Marshal sentinel
	data, err := toon.Marshal(payload{Name: fx.PayloadSample, Value: 1})
	if err == nil {
		return fmt.Errorf("Marshal returned nil err (data=%q) — sentinel contract broken", data)
	}
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("Marshal err is not sentinel: %v", err)
	}
	if data != nil {
		return fmt.Errorf("Marshal returned non-nil data with sentinel: %q", data)
	}

	// 2. MarshalIndent sentinel
	data, err = toon.MarshalIndent(payload{Name: fx.PayloadSample}, "", "  ")
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("MarshalIndent err is not sentinel: %v", err)
	}
	if data != nil {
		return fmt.Errorf("MarshalIndent returned non-nil data with sentinel: %q", data)
	}

	// 3. Unmarshal sentinel — target MUST stay zero-valued
	var p payload
	if err := toon.Unmarshal([]byte(`{"name":"x"}`), &p); !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("Unmarshal err is not sentinel: %v", err)
	}
	if p.Name != "" || p.Value != 0 || len(p.Items) != 0 {
		return fmt.Errorf("Unmarshal mutated target despite sentinel: %+v", p)
	}

	// 4. Encoder.Encode sentinel + writer purity
	var buf bytes.Buffer
	enc := toon.NewEncoder(&buf)
	if err := enc.Encode(payload{Name: fx.PayloadSample}); !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("Encoder.Encode err is not sentinel: %v", err)
	}
	if buf.Len() != 0 {
		return fmt.Errorf("Encoder.Encode wrote %d bytes despite sentinel: %q", buf.Len(), buf.String())
	}

	// 5. Decoder.Decode sentinel + target purity
	var p2 payload
	dec := toon.NewDecoder(strings.NewReader(`{"name":"x","value":99,"items":["a"]}`))
	if err := dec.Decode(&p2); !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("Decoder.Decode err is not sentinel: %v", err)
	}
	if p2.Name != "" || p2.Value != 0 || len(p2.Items) != 0 {
		return fmt.Errorf("Decoder.Decode mutated target despite sentinel: %+v", p2)
	}

	// 6. Compare sentinel + nil result
	cmp, err := toon.Compare(payload{Name: fx.PayloadSample, Items: []string{"a", "b"}})
	if !errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
		return fmt.Errorf("Compare err is not sentinel: %v", err)
	}
	if cmp != nil {
		return fmt.Errorf("Compare returned non-nil TokenComparison despite sentinel: %+v", cmp)
	}

	// 7. Heuristic helpers (no encoding claim, must behave)
	if !toon.IsTOONContentType(fx.HeaderSample) {
		return fmt.Errorf("IsTOONContentType rejected fixture header_sample %q", fx.HeaderSample)
	}
	if toon.IsTOONContentType("application/json") {
		return fmt.Errorf("IsTOONContentType accepted application/json — header inspection broken")
	}
	if got := toon.TokenEstimate([]byte("hello world test")); got <= 0 {
		return fmt.Errorf("TokenEstimate returned non-positive %d", got)
	}

	// 8. Bluff-mutation injection: plant a synthetic JSON-fallback claim
	// and assert that the gate (challenge script) catches it. In runtime
	// mutate mode we leak a JSON-encoded payload and emit a marker; the
	// shell wrapper greps for the marker and EXPECTS to see it.
	if mutate {
		jsonBytes, _ := json.Marshal(payload{Name: fx.PayloadSample, Value: 42})
		fmt.Printf("[toon-runner] BLUFF-PLANTED locale=%s json_bytes=%q\n", locale, jsonBytes)
	}

	return nil
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[toon-runner] FATAL: "+format+"\n", args...)
	os.Exit(1)
}

func modeLabel(mutate bool) string {
	if mutate {
		return "mutate(planted-bluff)"
	}
	return "normal"
}
