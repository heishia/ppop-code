import subprocess
import sys
import shutil
from pathlib import Path

VIVE_ROOT = Path(__file__).parent.parent
CONFIG_DIR = VIVE_ROOT / "config"


def test_config_directory_exists():
    assert CONFIG_DIR.exists(), f"Config directory not found: {CONFIG_DIR}"


def test_bridges_config_exists():
    bridges_path = CONFIG_DIR / "bridges.yaml"
    assert bridges_path.exists(), f"bridges.yaml not found: {bridges_path}"


def test_ppopcode_config_exists():
    config_path = CONFIG_DIR / "ppopcode.yaml"
    assert config_path.exists(), f"ppopcode.yaml not found: {config_path}"


def test_cursor_agent_available():
    cursor_path = shutil.which("cursor-agent")
    if cursor_path:
        print(f"cursor-agent found: {cursor_path}")
        return True
    else:
        print("WARNING: cursor-agent not found in PATH")
        print("Install from: https://docs.cursor.com/cli/installation")
        return False


def test_claude_cli_available():
    claude_path = shutil.which("claude")
    if claude_path:
        print(f"claude found: {claude_path}")
        return True
    else:
        print("WARNING: claude CLI not found in PATH")
        return False


def test_bridges_yaml_valid():
    import yaml
    
    bridges_path = CONFIG_DIR / "bridges.yaml"
    content = bridges_path.read_text(encoding="utf-8")
    
    try:
        config = yaml.safe_load(content)
        assert "cursor" in config, "Missing cursor section"
        assert "claude" in config, "Missing claude section"
        print("bridges.yaml is valid YAML")
    except yaml.YAMLError as e:
        raise AssertionError(f"Invalid YAML: {e}")


if __name__ == "__main__":
    test_config_directory_exists()
    test_bridges_config_exists()
    test_ppopcode_config_exists()
    
    print("\n--- CLI Availability Check ---")
    cursor_ok = test_cursor_agent_available()
    claude_ok = test_claude_cli_available()
    
    print("\n--- Config Validation ---")
    try:
        test_bridges_yaml_valid()
    except ImportError:
        print("SKIP: PyYAML not installed")
    
    print("\n--- Summary ---")
    print(f"Cursor CLI: {'OK' if cursor_ok else 'NOT FOUND'}")
    print(f"Claude CLI: {'OK' if claude_ok else 'NOT FOUND'}")
    print("Bridge tests completed!")
