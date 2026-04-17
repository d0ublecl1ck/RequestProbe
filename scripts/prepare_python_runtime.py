#!/usr/bin/env python3

from __future__ import annotations

import argparse
import shutil
import sys
from pathlib import Path


def runtime_site_packages(runtime_root: Path) -> Path:
    version = f"python{sys.version_info.major}.{sys.version_info.minor}"
    if sys.platform == "win32":
        return runtime_root / "Lib" / "site-packages"
    return runtime_root / "lib" / version / "site-packages"


def source_site_packages() -> Path:
    version = f"python{sys.version_info.major}.{sys.version_info.minor}"
    if sys.platform == "win32":
        return Path(sys.prefix) / "Lib" / "site-packages"
    return Path(sys.prefix) / "lib" / version / "site-packages"


def copy_tree(src: Path, dest: Path) -> None:
    if not src.exists():
        raise FileNotFoundError(f"source path does not exist: {src}")
    shutil.copytree(src, dest, symlinks=False)


def overlay_site_packages(src: Path, dest: Path) -> None:
    dest.mkdir(parents=True, exist_ok=True)
    for child in src.iterdir():
        target = dest / child.name
        if target.exists():
            if target.is_dir() and not target.is_symlink():
                shutil.rmtree(target)
            else:
                target.unlink()
        if child.is_dir() and not child.is_symlink():
            shutil.copytree(child, target, symlinks=False)
        else:
            shutil.copy2(child, target, follow_symlinks=True)


def build_runtime(dest: Path) -> None:
    base_prefix = Path(sys.base_prefix).resolve()
    venv_prefix = Path(sys.prefix).resolve()
    src_site = source_site_packages()

    if base_prefix == venv_prefix:
        raise RuntimeError("prepare_python_runtime.py must run inside a virtual environment created from the target runtime")

    if dest.exists():
        shutil.rmtree(dest)

    copy_tree(base_prefix, dest)
    overlay_site_packages(src_site, runtime_site_packages(dest))


def main() -> int:
    parser = argparse.ArgumentParser(description="Build a bundled Python runtime from the active virtual environment")
    parser.add_argument("--dest", required=True, help="Destination directory for the bundled runtime")
    args = parser.parse_args()

    build_runtime(Path(args.dest).resolve())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
