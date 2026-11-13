"""Fixture package: one function and one dependency, enough to install, test, audit and build."""

from packaging.version import Version


def hello(name: str) -> str:
    return f"Hello, {name}!"


def newer(a: str, b: str) -> str:
    return max(a, b, key=Version)
