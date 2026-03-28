#!/usr/bin/env python3
"""Patch spec-kitty merge preflight: do not require worktrees for lane=planned.

Upstream preflight treats every WP in the dependency graph as needing a worktree,
except lane=done. That blocks partial merges while WP04+ are still planned.

After `uv tool upgrade spec-kitty-cli`, re-run from the repo root:

  ./.kittify/scripts/repair-spec-kitty-merge-preflight.py
"""
from __future__ import annotations

import re
import shutil
import subprocess
import sys
from pathlib import Path


def _interpreter_from_spec_kitty() -> str | None:
    exe = shutil.which("spec-kitty")
    if not exe:
        return None
    line = Path(exe).read_text(encoding="utf-8", errors="replace").splitlines()[:1]
    if not line:
        return None
    m = re.match(r"^#!\s*(.+)$", line[0].strip())
    return m.group(1).strip() if m else None


def resolve_preflight_path() -> Path | None:
    interp = _interpreter_from_spec_kitty()
    candidates = [interp] if interp else []
    candidates.append(sys.executable)
    for py in candidates:
        if not py:
            continue
        proc = subprocess.run(
            [py, "-c", "import specify_cli.merge.preflight as p; print(p.__file__)"],
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            check=False,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return Path(proc.stdout.strip())
    return None


def main() -> int:
    path = resolve_preflight_path()
    if path is None:
        print(
            "Could not locate specify_cli.merge.preflight (install spec-kitty / spec-kitty-cli).",
            file=sys.stderr,
        )
        return 2

    text = path.read_text(encoding="utf-8")
    marker = "lane=planned; not started"
    if marker in text:
        print(f"OK (already patched): {path}")
        return 0

    needle = """            if lane == "done":
                result.warnings.append(
                    f"Skipping missing worktree check for {wp_id} (lane=done)."
                )
                continue

            result.passed = False"""

    insert = """            if lane == "done":
                result.warnings.append(
                    f"Skipping missing worktree check for {wp_id} (lane=done)."
                )
                continue

            # Planned WPs do not have worktrees until `spec-kitty agent workflow implement`;
            # they must not block merging completed WPs into the target branch.
            if lane == "planned":
                result.warnings.append(
                    f"Skipping missing worktree check for {wp_id} (lane=planned; not started)."
                )
                continue

            result.passed = False"""

    if needle not in text:
        print(
            f"Patch anchor not found in {path}; spec-kitty version may differ. Edit manually.",
            file=sys.stderr,
        )
        return 1

    path.write_text(text.replace(needle, insert, 1), encoding="utf-8")
    print(f"Patched: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
