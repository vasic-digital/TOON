# TOON Test Coverage Ledger

**CONST-050(B) requirement**: every exported symbol in `pkg/toon/toon.go`
MUST appear below with at least one captured-evidence test row.
Symbols added without a corresponding ledger row are a §11.4.30 /
CONST-055 audit failure.

**Round-286 baseline** (2026-05-19): all 11 exported symbols covered.

## Symbol → Test ledger

| Exported symbol                       | Test file / Challenge                                                    | Test name                              | Evidence type                                     |
|---------------------------------------|--------------------------------------------------------------------------|----------------------------------------|---------------------------------------------------|
| `ContentType` (const)                 | `pkg/toon/toon_test.go`                                                  | `TestContentType`                      | exact equality vs `"application/toon"`            |
| `ErrTOONEncodingNotImplemented` (var) | `pkg/toon/toon_test.go`                                                  | `TestMarshal_ReturnsSentinel` (+5 more)| `errors.Is` assertion across every entry point   |
| `Encoder` (type)                      | `pkg/toon/toon_test.go`                                                  | `TestEncoder_ReturnsSentinel`          | constructor + sentinel propagation                |
| `NewEncoder` (func)                   | `pkg/toon/toon_test.go`                                                  | `TestEncoder_ReturnsSentinel`          | constructor returns non-nil                       |
| `Encoder.Encode` (method)             | `pkg/toon/toon_test.go`                                                  | `TestEncoder_ReturnsSentinel`          | sentinel + writer-purity (`buf.Len() == 0`)       |
| `Decoder` (type)                      | `pkg/toon/toon_test.go`                                                  | `TestDecoder_ReturnsSentinel`          | constructor + sentinel propagation                |
| `NewDecoder` (func)                   | `pkg/toon/toon_test.go`                                                  | `TestDecoder_ReturnsSentinel`          | constructor returns non-nil                       |
| `Decoder.Decode` (method)             | `pkg/toon/toon_test.go`                                                  | `TestDecoder_ReturnsSentinel`          | sentinel + target purity                          |
| `Marshal` (func)                      | `pkg/toon/toon_test.go`                                                  | `TestMarshal_ReturnsSentinel`          | sentinel + nil data                               |
| `MarshalIndent` (func)                | `pkg/toon/toon_test.go`                                                  | `TestMarshalIndent_ReturnsSentinel`    | sentinel + nil data                               |
| `Unmarshal` (func)                    | `pkg/toon/toon_test.go`                                                  | `TestUnmarshal_ReturnsSentinel`        | sentinel + target-untouched                       |
| `IsTOONContentType` (func)            | `pkg/toon/toon_test.go`                                                  | `TestIsTOONContentType`                | header match + case-insensitivity                 |
| `TokenEstimate` (func)                | `pkg/toon/toon_test.go`                                                  | `TestTokenEstimate`                    | positive integer result                           |
| `TokenComparison` (type)              | `pkg/toon/toon_test.go`                                                  | `TestCompare_ReturnsSentinel`          | nil return on sentinel                            |
| `Compare` (func)                      | `pkg/toon/toon_test.go`                                                  | `TestCompare_ReturnsSentinel`          | sentinel + nil comparison                         |

## Challenge ledger (CONST-050(B) integration / E2E / paired-mutation)

| Challenge file                                                  | What it proves                                                                                      | Normal exit | Mutate exit |
|-----------------------------------------------------------------|-----------------------------------------------------------------------------------------------------|-------------|-------------|
| `challenges/runner/main.go`                                     | Real Go binary exercises every public symbol in 5 locales (en, sr, ja, es, de); zero JSON bytes    | 0           | n/a         |
| `challenges/scripts/toon_describe_challenge.sh`                 | Builds + invokes runner; asserts sentinel propagation; CONST-055 paired-mutation gate              | 0           | 99          |
| `challenges/scripts/no_suspend_calls_challenge.sh`              | CONST-033 host-power-management source scan                                                         | 0           | n/a         |
| `challenges/scripts/host_no_auto_suspend_challenge.sh`          | CONST-033 host configuration verification                                                           | 0           | n/a         |
| `challenges/scripts/chaos_failure_injection_challenge.sh`       | CONST-050(B) chaos test type                                                                        | 0           | n/a         |
| `challenges/scripts/ddos_health_flood_challenge.sh`             | CONST-050(B) DDoS test type                                                                         | 0           | n/a         |
| `challenges/scripts/scaling_horizontal_challenge.sh`            | CONST-050(B) scaling test type                                                                      | 0           | n/a         |
| `challenges/scripts/stress_sustained_load_challenge.sh`         | CONST-050(B) stress test type                                                                       | 0           | n/a         |
| `challenges/scripts/ui_terminal_interaction_challenge.sh`       | CONST-050(B) UI test type                                                                           | 0           | n/a         |
| `challenges/scripts/ux_end_to_end_flow_challenge.sh`            | CONST-050(B) UX test type                                                                           | 0           | n/a         |

## Captured runtime evidence (round-286 baseline)

```
$ go test -race ./...
ok      digital.vasic.toon/pkg/toon     1.008s

$ bash challenges/scripts/toon_describe_challenge.sh
[toon-describe-challenge] building runner...
[toon-describe-challenge] locale en — sentinel OK
[toon-describe-challenge] locale sr — sentinel OK
[toon-describe-challenge] locale ja — sentinel OK
[toon-describe-challenge] locale es — sentinel OK
[toon-describe-challenge] locale de — sentinel OK
[toon-describe-challenge] all 5 locales asserted sentinel; zero JSON bytes leaked
[toon-describe-challenge] PASS (exit 0)

$ bash challenges/scripts/toon_describe_challenge.sh --anti-bluff-mutate
[toon-describe-challenge] mutate mode — planting JSON-fallback bluff
[toon-describe-challenge] anti-bluff gate caught planted bluff: FAIL as designed
[toon-describe-challenge] mutate mode exit 99 (proves the gate actually catches bluffs)
```

## Coverage gaps + future work

When native TOON encoding is wired (upstream `toon-format/toon-go` or
in-repo hand-written encoder), the ledger MUST be extended in the
SAME commit as the implementation:

- Add round-positive tests asserting actual TOON byte output (not
  JSON, not sentinel) — token savings vs JSON measured and asserted
  against a non-trivial floor (e.g. ≥30%).
- Extend `challenges/runner/main.go` to perform round-trip encode →
  decode and assert byte-equality of payload.
- Update `toon_describe_challenge.sh` paired-mutation to plant a
  bogus encoder and assert detection.

A ledger row added without a corresponding captured-evidence test
is itself a §11.4 PASS-bluff at the documentation layer.

## Constitutional anchors

- CONST-035 — anti-bluff posture
- CONST-048 — full-automation-coverage mandate
- CONST-050(A) — no fakes beyond unit tests
- CONST-050(B) — 100% test-type coverage (this document is the ledger)
- CONST-055 — post-constitution-pull validation
- Article XI §11.9 — anti-bluff forensic anchor
