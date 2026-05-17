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

That bluff has been removed. Every entry point now returns the
sentinel error `toon.ErrTOONEncodingNotImplemented`. There is no
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

## Future work

Native TOON encoding will be wired once one of the following is
available and integration is approved:

- upstream `toon-format/toon-go` library (the project's reference
  implementation upstream), OR
- an in-repo hand-written TOON encoder that delivers the format's
  advertised token-efficiency characteristics.

Until then, **no claim of token savings is made by this module**.

## Constitutional anchors

- CONST-035 — anti-bluff posture
- CONST-050(A) — no fakes beyond unit tests
- Article XI §11.9 — anti-bluff forensic anchor
