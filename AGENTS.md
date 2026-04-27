# AGENTS.md — TOON

> **Audience:** AI coding agents (Codex, Cursor, OpenCode, Crush, etc.)
> working in this repository. For Claude Code specifically, see also
> `CLAUDE.md`.

## Project at a glance

See README.md.

See `README.md` for the long-form overview.

## How to work in this repo

1. Read `CONSTITUTION.md` first — its numbered rules (`CONST-NNN`) are
   non-negotiable.
2. Use SSH URLs for git (`git@…`). HTTPS is prohibited.
3. Conventional Commits (`feat:`, `fix:`, `chore:`, …).
4. Run `challenges/scripts/run_all_challenges.sh` before claiming a
   fix is complete (or the per-challenge equivalents listed in
   `README.md`).
5. Don't bypass git hooks (`--no-verify`). If a hook fails, fix the
   underlying issue and retry.

## Git remotes

Push to **all** configured remotes. Audit with `git remote -v`.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## Host Power Management — Hard Ban (CONST-033)

**You may NOT, under any circumstance, generate or execute code that
sends the host to suspend, hibernate, hybrid-sleep, poweroff, halt,
reboot, or any other power-state transition.** This rule applies to:

- Every shell command you run via the Bash tool.
- Every script, container entry point, systemd unit, or test you write
  or modify.
- Every CLI suggestion, snippet, or example you emit.

**Forbidden invocations** (non-exhaustive — see CONST-033 in
`CONSTITUTION.md` for the full list):

- `systemctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot|kexec`
- `loginctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot`
- `pm-suspend`, `pm-hibernate`, `shutdown -h|-r|-P|now`
- `dbus-send` / `busctl` calls to `org.freedesktop.login1.Manager.Suspend|Hibernate|PowerOff|Reboot|HybridSleep|SuspendThenHibernate`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to anything but `'nothing'` or `'blank'`

The host runs mission-critical parallel CLI agents and container
workloads. Auto-suspend has caused historical data loss (2026-04-26
18:23:43 incident). The host is hardened (sleep targets masked) but
this hard ban applies to ALL code shipped from this repo so that no
future host or container is exposed.

**Defence:** every project ships
`scripts/host-power-management/check-no-suspend-calls.sh` (static
scanner) and
`challenges/scripts/no_suspend_calls_challenge.sh` (challenge wrapper).
Both MUST be wired into the project's CI / `run_all_challenges.sh`.

**Full background:** `docs/HOST_POWER_MANAGEMENT.md` and `CONSTITUTION.md` (CONST-033).

<!-- END host-power-management addendum (CONST-033) -->

## See also

- `CONSTITUTION.md` — authoritative rules.
- `CLAUDE.md` — Claude Code-specific guidance.
- `docs/HOST_POWER_MANAGEMENT.md` — CONST-033 background and runbook.
