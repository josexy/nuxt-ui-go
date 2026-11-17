#!/usr/bin/env python3
"""Check formatting, MyGo version mentions, tests, lifecycle rules, generated sources and cross builds.

Runs on Windows, Linux and macOS with Python 3.9 or later and only the
standard library. Use `task verify`, or run it directly from any directory:

    python3 scripts/verify.py
"""

from __future__ import annotations

import argparse
import ast
import hashlib
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

MIN_PYTHON = (3, 9)
ROOT = Path(__file__).resolve().parent.parent
CROSS_TARGETS = (("windows", "amd64"), ("linux", "amd64"), ("darwin", "arm64"))
# The full Lucide set exceeds the Windows command-line length limit in one call.
GOFMT_BATCH = 100
# go.mod is the one place that names the MyGo version: generators read it from
# there, and docs and comments state current behavior without a version, backed
# by tests that fail when MyGo changes it. A version written anywhere else goes
# stale with the next upgrade. "Since MyGo ..." history and CHANGELOG.md may
# name one.
MYGO_VERSION = re.compile(r"(?i)\bmygo\b[^\w\n]{0,3}(?:tree/|releases/tag/)?v?\d+\.\d+\.\d+")
SINCE = re.compile(r"(?i)\bsince\s+$")
VERSION_SCAN_FILES = ("AGENTS.md", "README.md", "THIRD_PARTY_NOTICES.md", "Taskfile.yml")
VERSION_SCAN_DIRS = (".agents", ".github", "docs", "examples", "internal", "scripts", "ui", "website")
VERSION_SCAN_SUFFIXES = {".go", ".js", ".json", ".md", ".mjs", ".py", ".ts", ".vue", ".yaml", ".yml"}
# Dependencies, build output and generated files, which take the version from go.mod.
VERSION_SCAN_SKIP = {
    ".agents/skills/nuxt-ui-go/references/components",
    "website/.data",
    "website/.nuxt",
    "website/.output",
    "website/app/data",
    "website/content",
    "website/dist",
    "website/node_modules",
    "website/public/preview",
}


class CheckFailed(Exception):
    pass


def tool(name: str) -> str:
    path = shutil.which(name)
    if path is None:
        raise CheckFailed(f"{name} was not found on PATH")
    return path


def run(*args: str, env: dict[str, str] | None = None, capture: bool = False) -> str:
    command = [tool(args[0]), *args[1:]]
    result = subprocess.run(
        command,
        cwd=ROOT,
        env=env,
        stdout=subprocess.PIPE if capture else None,
        text=capture,
    )
    if result.returncode != 0:
        raise CheckFailed(f"{' '.join(args)} exited with status {result.returncode}")
    return result.stdout if capture else ""


def check_python_scripts() -> None:
    # CI runs a newer Python, so parse with the oldest supported grammar.
    for script in sorted((ROOT / "scripts").glob("*.py")):
        source = script.read_text(encoding="utf-8")
        try:
            ast.parse(source, str(script), feature_version=MIN_PYTHON)
        except SyntaxError as err:
            raise CheckFailed(f"{script.relative_to(ROOT)}: {err}") from err


def check_gofmt() -> None:
    files = [p for p in ROOT.glob("*.go") if p.is_file()]
    for directory in ("ui", "icons", "internal", "examples"):
        files.extend(p for p in (ROOT / directory).rglob("*.go") if p.is_file())
    names = sorted(str(p.relative_to(ROOT)) for p in files)
    for offset in range(0, len(names), GOFMT_BATCH):
        unformatted = run("gofmt", "-l", *names[offset : offset + GOFMT_BATCH], capture=True)
        if unformatted.strip():
            raise CheckFailed(f"gofmt required:\n{unformatted.rstrip()}")


def version_scan_files() -> list[Path]:
    files = [ROOT / name for name in VERSION_SCAN_FILES if (ROOT / name).is_file()]
    for directory in VERSION_SCAN_DIRS:
        for current, dirs, names in os.walk(ROOT / directory):
            here = Path(current)
            dirs[:] = sorted(d for d in dirs if (here / d).relative_to(ROOT).as_posix() not in VERSION_SCAN_SKIP)
            files.extend(here / n for n in sorted(names) if Path(n).suffix in VERSION_SCAN_SUFFIXES)
    return files


def check_mygo_version_mentions() -> None:
    found = []
    for path in version_scan_files():
        text = path.read_text(encoding="utf-8", errors="replace")
        for number, line in enumerate(text.splitlines(), 1):
            for match in MYGO_VERSION.finditer(line):
                if not SINCE.search(line[: match.start()]):
                    found.append(f"  {path.relative_to(ROOT).as_posix()}:{number}: {match.group(0)}")
    if found:
        listing = "\n".join(found)
        raise CheckFailed(
            "MyGo versions outside go.mod go stale with the next upgrade. Read the version from go.mod,"
            " or state the behavior without one and keep a test that fails when MyGo changes it:\n" + listing
        )


def generated_fingerprint() -> dict[str, str]:
    # Key by file name as well as content, so additions and removals also fail.
    files = [p for p in (ROOT / "icons" / "lucide").glob("*.go") if p.is_file()]
    palette = ROOT / "ui" / "palette_gen.go"
    if palette.is_file():
        files.append(palette)
    return {
        p.relative_to(ROOT).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest()
        for p in files
    }


def check_generated() -> None:
    before = generated_fingerprint()
    run("go", "generate", "./...")
    after = generated_fingerprint()
    drift = sorted(name for name in before.keys() | after.keys() if before.get(name) != after.get(name))
    if drift:
        listing = "\n".join(f"  {name}" for name in drift)
        raise CheckFailed(f"Generated sources changed; regenerate and include the updated files:\n{listing}")


def cross_build() -> None:
    for goos, goarch in CROSS_TARGETS:
        run("go", "build", "./...", env={**os.environ, "GOOS": goos, "GOARCH": goarch})
        print(f"Built {goos}/{goarch} with CGO_ENABLED=0", flush=True)


def main() -> int:
    argparse.ArgumentParser(description=__doc__.splitlines()[0]).parse_args()
    if sys.version_info < MIN_PYTHON:
        print("Python {}.{} or later is required.".format(*MIN_PYTHON), file=sys.stderr)
        return 1
    os.environ["CGO_ENABLED"] = "0"
    os.environ.pop("GOOS", None)
    os.environ.pop("GOARCH", None)
    try:
        check_python_scripts()
        check_gofmt()
        check_mygo_version_mentions()
        run("go", "test", "./...", "-count=1")
        run("go", "vet", "./...")
        run("go", "tool", "mygo", "vet", ".")
        check_generated()
        cross_build()
    except CheckFailed as err:
        print(err, file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        return 130
    return 0


if __name__ == "__main__":
    sys.exit(main())
