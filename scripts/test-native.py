#!/usr/bin/env python3
"""Run the native Gallery interaction tests in a desktop session.

Runs on Windows, Linux and macOS with Python 3.9 or later and only the
standard library. Use `task test:native` or `task test:native:smoke`, or run
it directly from any directory:

    python3 scripts/test-native.py [--hidden-smoke]

Screenshots and logs are written to .work/native/.
"""

from __future__ import annotations

import argparse
import contextlib
import os
import shutil
import signal
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Iterator

MIN_PYTHON = (3, 9)
ROOT = Path(__file__).resolve().parent.parent
ARTIFACT_DIR = ROOT / ".work" / "native"
WINDOWS = sys.platform == "win32"
NARROW_WIDTH = 720


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run the native Gallery interaction tests. Requires a desktop session; "
        "on Linux, DISPLAY or WAYLAND_DISPLAY must be set, even for the hidden smoke test.",
    )
    parser.add_argument(
        "--hidden-smoke",
        action="store_true",
        help="only check the first frame of a hidden window at the default width",
    )
    return parser.parse_args()


def tool(name: str) -> str:
    path = shutil.which(name)
    if path is None:
        sys.exit(f"{name} was not found on PATH")
    return path


@contextlib.contextmanager
def parked_pointer() -> Iterator[None]:
    """Move the Windows cursor to a corner for the run, then put it back.

    Input injection does not move the OS cursor. A Windows mouse move at its
    actual position can otherwise interrupt an injected drag or hover sequence.
    """
    if not WINDOWS:
        yield
        return
    import ctypes
    from ctypes import wintypes

    user32 = ctypes.WinDLL("user32")
    user32.GetCursorPos.argtypes = [ctypes.POINTER(wintypes.POINT)]
    user32.GetCursorPos.restype = wintypes.BOOL
    user32.SetCursorPos.argtypes = [ctypes.c_int, ctypes.c_int]
    user32.SetCursorPos.restype = wintypes.BOOL
    saved = wintypes.POINT()
    restore = bool(user32.GetCursorPos(ctypes.byref(saved)))
    try:
        user32.SetCursorPos(0, 0)
        yield
    finally:
        if restore:
            user32.SetCursorPos(saved.x, saved.y)


def make_socket_dir() -> Path:
    if WINDOWS:
        # Keep the socket on the checkout filesystem: redirected Windows TEMP
        # directories can create AF_UNIX endpoints that clients cannot connect to.
        parent = ARTIFACT_DIR
    else:
        # Keep the Unix socket path short even when the checkout has a long path.
        parent = Path("/tmp")
    return Path(tempfile.mkdtemp(prefix="nui-native.", dir=parent))


def stop(process: subprocess.Popen[bytes]) -> None:
    process.terminate()
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait()


def start_gallery(gallery: Path, run_dir: Path, width: int, hidden: bool) -> subprocess.Popen[bytes]:
    args = [str(gallery), "-static", "-width", str(width)]
    if hidden:
        args.append("-hidden")
    # Do not open a console window for the Gallery on Windows.
    flags = subprocess.CREATE_NO_WINDOW if WINDOWS else 0
    with open(run_dir / "stdout.log", "wb") as stdout, open(run_dir / "stderr.log", "wb") as stderr:
        return subprocess.Popen(args, cwd=ROOT, stdout=stdout, stderr=stderr, creationflags=flags)


def run_suite(hidden: bool) -> int:
    go = tool("go")
    ARTIFACT_DIR.mkdir(parents=True, exist_ok=True)
    gallery = ARTIFACT_DIR / ("gallery-e2e.exe" if WINDOWS else "gallery-e2e")
    build = subprocess.run([go, "build", "-tags", "e2e", "-o", str(gallery), "./examples/gallery"], cwd=ROOT)
    if build.returncode != 0:
        print("Gallery build failed", file=sys.stderr)
        return 1

    socket_dir = make_socket_dir()
    socket_path = socket_dir / "gallery.sock"
    os.environ["MYGO_AUTOMATION"] = str(socket_path)
    os.environ["NUI_AUTOMATION_SOCKET"] = str(socket_path)
    process = None
    try:
        for width in (1080,) if hidden else (1080, NARROW_WIDTH):
            run_dir = ARTIFACT_DIR
            os.environ["NUI_NATIVE_NARROW"] = "0"
            if width == NARROW_WIDTH:
                run_dir = ARTIFACT_DIR / "narrow"
                os.environ["NUI_NATIVE_NARROW"] = "1"
                run_dir.mkdir(exist_ok=True)
            os.environ["NUI_NATIVE_ARTIFACTS"] = str(run_dir)
            process = start_gallery(gallery, run_dir, width, hidden)
            probe = subprocess.run(
                [go, "test", "./internal/nativeprobe", "-run", "^TestNativeGallery$", "-v", "-count=1"],
                cwd=ROOT,
            )
            if probe.returncode != 0:
                print(f"Native validation failed; see {run_dir}", file=sys.stderr)
                return 1
            stop(process)
            process = None
            socket_path.unlink(missing_ok=True)
    finally:
        if process is not None:
            stop(process)
        # Only these two exact paths, created by this invocation; no recursive deletion.
        socket_path.unlink(missing_ok=True)
        try:
            socket_dir.rmdir()
        except OSError as err:
            print(f"Could not remove {socket_dir}: {err}", file=sys.stderr)
    return 0


def main() -> int:
    args = parse_args()
    if sys.version_info < MIN_PYTHON:
        print("Python {}.{} or later is required.".format(*MIN_PYTHON), file=sys.stderr)
        return 1
    if sys.platform.startswith("linux") and not (os.environ.get("DISPLAY") or os.environ.get("WAYLAND_DISPLAY")):
        print("Native tests require a Linux desktop session (DISPLAY or WAYLAND_DISPLAY).", file=sys.stderr)
        return 1
    # Run the cleanup in run_suite when the task runner or CI stops the script.
    signal.signal(signal.SIGTERM, lambda signum, frame: sys.exit(128 + signum))

    os.environ["CGO_ENABLED"] = "0"
    os.environ.pop("GOOS", None)
    os.environ.pop("GOARCH", None)
    # Verify Gallery falls back to English even with a Chinese language preference.
    os.environ["MYGO_LOCALE"] = "zh-Hans"
    os.environ["NUI_NATIVE_HIDDEN"] = "1" if args.hidden_smoke else "0"
    try:
        with parked_pointer():
            return run_suite(args.hidden_smoke)
    except KeyboardInterrupt:
        return 130


if __name__ == "__main__":
    sys.exit(main())
