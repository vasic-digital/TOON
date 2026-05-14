# CLAUDE.md — TOON

This file provides guidance to Claude Code (claude.ai/code) when
working with code in this repository.

## Read first

1. `CONSTITUTION.md` — numbered, non-negotiable rules (`CONST-NNN`).
2. `AGENTS.md` — generic AI-agent guidance.
3. This file — Claude-specific reminders.

## Project at a glance

See README.md.

See `README.md` for the long-form overview.

## Hard stops (non-negotiable)

1. **No HTTPS git URLs.** SSH only (`git@…`).
2. **No `--no-verify`.** Hooks always run; if a hook fails, fix the
   underlying issue and retry.
3. **Conventional Commits.**
4. **All challenges PASS** (`bash challenges/scripts/run_all_challenges.sh`
   or per-challenge equivalents) before claiming a fix is complete.
5. **Definition of Done:** see `CONSTITUTION.md` — pasted terminal
   output as evidence; no self-certification words without it.

## Git remotes

Push to **all** configured remotes. Audit with `git remote -v`.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->

## See also

- `CONSTITUTION.md` — authoritative rules.
- `AGENTS.md` — guidance for non-Claude agents.
- `docs/HOST_POWER_MANAGEMENT.md` — CONST-033 background and runbook.

---

## Article XI §11.9 — Anti-Bluff Forensic Anchor (cascaded from parent CONSTITUTION.md)

> Verbatim user mandate (2026-04-29, reasserted multiple times across 2026-05): *"We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completion and full usability by end users of the product!"*

Operative rule: **The bar for shipping is not "tests pass" but "users can use the feature."** Every PASS in this codebase MUST carry positive runtime evidence captured during execution. Metadata-only / configuration-only / absence-of-error / grep-based PASS without runtime evidence are critical defects regardless of how green the summary line looks. No false-success results are tolerable.

This anchor MUST remain in this submodule's CONSTITUTION.md, CLAUDE.md, and AGENTS.md alongside CONST-047 — see the parent repository's `CONSTITUTION.md` for the full text.
