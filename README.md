# digital.vasic.toon

Generic reusable Go module **intended** to expose a Token-Oriented
Object Notation (TOON) encoding/decoding API. TOON is designed to be
a compact, human-readable serialization format optimized for LLM token
efficiency.

## STATUS: PENDING_IMPLEMENTATION (round-27 §11.4 audit, 2026-05-17)

Native TOON encoding is **NOT YET IMPLEMENTED** in this module.

Earlier revisions of `pkg/toon/toon.go` exposed `Marshal()`,
`Unmarshal()`, `MarshalIndent()`, `Encoder.Encode`, `Decoder.Decode`,
and `Compare()` while internally delegating to `encoding/json` and
documenting themselves as "TOON encoding" — a **CRITICAL contract
bluff** at the module-purpose layer. The `Compare()` helper went
further and computed `Savings: 0.0` from two identical `json.Marshal`
calls while presenting the result as if it had measured TOON vs JSON.

That bluff has been removed. Every encoding entry point now returns
the sentinel error `toon.ErrTOONEncodingNotImplemented`. There is no
silent JSON fallback — callers who need JSON must use `encoding/json`
directly.

## Installation

```bash
go get digital.vasic.toon
```

## Usage (current — sentinel-error behaviour)

```go
import (
    "errors"
    "digital.vasic.toon/pkg/toon"
)

data, err := toon.Marshal(myStruct)
if errors.Is(err, toon.ErrTOONEncodingNotImplemented) {
    // Expected today. Use encoding/json directly until native TOON
    // encoding is wired (see "Future work" below).
}
```

`toon.IsTOONContentType(header)` and `toon.TokenEstimate(data)` remain
useful as standalone helpers — they make no encoding claims and were
not part of the §11.4 bluff scope.

## Anti-Bluff Guarantees (round-286 §11.4 enrichment, 2026-05-19)

This module ships with positive-evidence guarantees per CONST-035 /
Article XI §11.9. Every public symbol's observable behaviour is
proven by a paired test or Challenge that captures runtime evidence:

1. **Sentinel-error contract** — `Marshal`, `Unmarshal`, `MarshalIndent`,
   `Encoder.Encode`, `Decoder.Decode`, and `Compare` ALL return
   `ErrTOONEncodingNotImplemented`. The contract is asserted via
   `errors.Is(err, ErrTOONEncodingNotImplemented)`, not by substring
   matching on the error message — the sentinel is the API.
2. **Encoder writer purity** — `Encoder.Encode` returning the sentinel
   MUST NOT write any bytes to the underlying `io.Writer`. The
   `TestEncoder_ReturnsSentinel` test asserts `buf.Len() == 0` on
   sentinel return, so a future regression that "accidentally" writes
   JSON to the buffer would FAIL the test.
3. **Decoder target purity** — `Decoder.Decode` returning the sentinel
   MUST NOT populate the target value. Tests assert the target
   remains zero-valued.
4. **No silent JSON fallback** — anti-bluff Challenge
   (`challenges/runner/main.go` + `toon_describe_challenge.sh`)
   exercises every public symbol from a real Go binary, captures the
   sentinel propagation through five locales (en, sr, ja, es, de),
   and asserts ZERO bytes were ever JSON-encoded in the process.
5. **Heuristic helpers honesty** — `IsTOONContentType` and
   `TokenEstimate` make NO TOON-specific claim. They are pure string
   inspection and `len/4` arithmetic; their tests verify exactly that
   minimal contract and nothing more.
6. **Paired-mutation gate** — `toon_describe_challenge.sh` runs in
   two modes: normal (exit 0) and `--anti-bluff-mutate` (exit 99).
   The mutate mode plants a contrived JSON-fallback scenario; the
   gate MUST detect the planted bluff and refuse. A passing mutate
   run is itself a CONST-055 violation.

## Future work

Native TOON encoding will be wired once one of the following is
available and integration is approved:

- upstream `toon-format/toon-go` library (the project's reference
  implementation upstream), OR
- an in-repo hand-written TOON encoder that delivers the format's
  advertised token-efficiency characteristics.

Until then, **no claim of token savings is made by this module**.
When integration lands, the §11.4 audit cycle will repeat: every
new encoding path MUST ship with captured runtime evidence proving
the advertised behaviour — no PASS without bytes-on-wire.

## Running the Challenges

```bash
# Normal anti-bluff run (must exit 0):
bash challenges/scripts/toon_describe_challenge.sh

# Paired-mutation run (must exit 99 — proves the gate actually catches bluffs):
bash challenges/scripts/toon_describe_challenge.sh --anti-bluff-mutate

# Direct runner invocation (5-locale bilingual sentinel walk):
cd challenges/runner && go run . -locales en,sr,ja,es,de
```

## Test Coverage

See `docs/test-coverage.md` for the symbol → test ledger
(CONST-050(B) requirement). Every exported symbol in
`pkg/toon/toon.go` MUST appear in the ledger with at least one
captured-evidence test row. Symbols added without ledger update are
a §11.4.30 audit failure.

## Constitutional anchors

- CONST-035 — anti-bluff posture
- CONST-047 — recursive submodule application
- CONST-048 — full-automation-coverage mandate
- CONST-050(A) — no fakes beyond unit tests
- CONST-050(B) — 100% test-type coverage
- CONST-053 — `.gitignore` + no-versioned-build-artifacts
- CONST-055 — post-constitution-pull validation
- Article XI §11.9 — anti-bluff forensic anchor
