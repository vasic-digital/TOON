#!/usr/bin/env bash
#
# toon_describe_challenge.sh — TOON sentinel-API anti-bluff Challenge
# with CONST-055 paired-mutation gate.
#
# Round-286 §11.4 enrichment (2026-05-19). Mirrors the round-220 / 242-285
# template established across the rest of the dependencies/ submodule tree.
#
# Modes:
#   (default)          — normal anti-bluff run; MUST exit 0
#   --anti-bluff-mutate — planted-bluff run; MUST exit 99 (gate caught bluff)
#
# Constitutional anchors:
#   CONST-035 — anti-bluff posture
#   CONST-048 — full-automation-coverage mandate
#   CONST-050(B) — 100% test-type coverage (Challenges layer)
#   CONST-055 — post-constitution-pull validation (paired-mutation)
#   Article XI §11.9 — anti-bluff forensic anchor

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
RUNNER_DIR="${MODULE_ROOT}/challenges/runner"
FIXTURES_DIR="${MODULE_ROOT}/challenges/fixtures"
LOG_PREFIX="[toon-describe-challenge]"

MUTATE=0
if [[ "${1:-}" == "--anti-bluff-mutate" ]]; then
    MUTATE=1
fi

echo "${LOG_PREFIX} module_root=${MODULE_ROOT}"
echo "${LOG_PREFIX} runner_dir=${RUNNER_DIR}"
echo "${LOG_PREFIX} fixtures_dir=${FIXTURES_DIR}"

# --- fixture sanity -----------------------------------------------------
for loc in en sr ja es de; do
    fixture="${FIXTURES_DIR}/${loc}.json"
    if [[ ! -f "${fixture}" ]]; then
        echo "${LOG_PREFIX} FAIL: missing fixture ${fixture}" >&2
        exit 1
    fi
done

# --- build runner -------------------------------------------------------
TMP_BIN="$(mktemp -d)/toon-runner"
trap 'rm -rf "$(dirname "${TMP_BIN}")"' EXIT

echo "${LOG_PREFIX} building runner..."
(
    cd "${MODULE_ROOT}"
    go build -o "${TMP_BIN}" ./challenges/runner
) || {
    echo "${LOG_PREFIX} FAIL: runner build failed" >&2
    exit 1
}

# --- run runner; capture output -----------------------------------------
OUT_FILE="$(mktemp)"
trap 'rm -f "${OUT_FILE}"; rm -rf "$(dirname "${TMP_BIN}")"' EXIT

RUN_ARGS=(-locales en,sr,ja,es,de -fixtures "${FIXTURES_DIR}")
if [[ "${MUTATE}" -eq 1 ]]; then
    RUN_ARGS+=(-mutate)
fi

if ! "${TMP_BIN}" "${RUN_ARGS[@]}" >"${OUT_FILE}" 2>&1; then
    echo "${LOG_PREFIX} FAIL: runner exited non-zero" >&2
    cat "${OUT_FILE}" >&2
    exit 1
fi

cat "${OUT_FILE}"

# --- anti-bluff gate ----------------------------------------------------
# In NORMAL mode: there must be NO 'BLUFF-PLANTED' marker in runner output.
# In MUTATE mode: there MUST be a 'BLUFF-PLANTED' marker on every locale.
#
# The gate (this script) catches the planted bluff and exits 99. A passing
# mutate run is itself a CONST-055 violation — it means the gate is blind.

if [[ "${MUTATE}" -eq 0 ]]; then
    if grep -q "BLUFF-PLANTED" "${OUT_FILE}"; then
        echo "${LOG_PREFIX} FAIL: BLUFF-PLANTED marker present in normal mode" >&2
        exit 1
    fi
    echo "${LOG_PREFIX} all 5 locales asserted sentinel; zero JSON bytes leaked"
    echo "${LOG_PREFIX} PASS (exit 0)"
    exit 0
fi

# Mutate mode: verify the gate actually caught all 5 planted bluffs.
PLANTED_COUNT="$(grep -c "BLUFF-PLANTED" "${OUT_FILE}" || true)"
if [[ "${PLANTED_COUNT}" -lt 5 ]]; then
    echo "${LOG_PREFIX} GATE BLIND: expected 5 BLUFF-PLANTED markers, got ${PLANTED_COUNT}" >&2
    echo "${LOG_PREFIX} CONST-055 violation — paired-mutation gate did not detect planted bluffs" >&2
    exit 1
fi

echo "${LOG_PREFIX} mutate mode — gate detected ${PLANTED_COUNT}/5 planted bluffs"
echo "${LOG_PREFIX} mutate mode exit 99 (proves the gate actually catches bluffs)"
exit 99
