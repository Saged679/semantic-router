"""The CLI's suite collects on every Python version the package supports."""

import ast
import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
CLI_ROOT = REPO_ROOT / "src" / "vllm-sr"
PYPROJECT = CLI_ROOT / "pyproject.toml"
# ``except*`` groups are a separate node from Python 3.11 on.
TRY_NODES = (ast.Try, getattr(ast, "TryStar", ast.Try))


def _tomllib_imports(path: Path) -> list[bool]:
    """Every tomllib import in the file, with whether a fallback guards it."""
    tree = ast.parse(path.read_text(encoding="utf-8"))
    imports: list[bool] = []

    def visit(node: ast.AST, guarded: bool) -> None:
        for child in ast.iter_child_nodes(node):
            if isinstance(child, TRY_NODES):
                # A tomllib import in the try body is guarded by the handlers.
                for statement in child.body:
                    visit(statement, True)
                continue
            if isinstance(child, ast.Import):
                for alias in child.names:
                    if alias.name == "tomllib":
                        imports.append(guarded)
            visit(child, guarded)

    visit(tree, False)
    return imports


def _sources() -> list[Path]:
    """The package's own Python files, without any environment beside them."""
    sources: list[Path] = []
    for path in sorted(CLI_ROOT.rglob("*.py")):
        parts = path.relative_to(CLI_ROOT).parts
        if "site-packages" in parts or "__pycache__" in parts:
            continue
        # A virtual environment beside the package carries its own copies.
        directory = CLI_ROOT
        for part in parts[:-1]:
            directory = directory / part
            if (directory / "pyvenv.cfg").exists():
                break
        else:
            sources.append(path)
    return sources


def test_the_package_declares_the_python_versions_it_supports() -> None:
    text = PYPROJECT.read_text(encoding="utf-8")

    assert 'requires-python = ">=3.10"' in text
    # The suite reads TOML, which the standard library only parses from
    # Python 3.11 on, so the backport has to be declared for 3.10.
    assert re.search(r"tomli[^,\]]*;\s*python_version\s*<\s*'3\.11'", text)


def test_every_tomllib_import_falls_back_for_python_310() -> None:
    for path in _sources():
        for guarded in _tomllib_imports(path):
            assert guarded, f"{path} imports tomllib without a Python 3.10 fallback"
