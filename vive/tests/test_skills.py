import subprocess
import sys
from pathlib import Path

VIVE_ROOT = Path(__file__).parent.parent
SKILLS_DIR = VIVE_ROOT / ".claude" / "skills"


def test_skills_directory_exists():
    assert SKILLS_DIR.exists(), f"Skills directory not found: {SKILLS_DIR}"


def test_cursor_edit_skill_exists():
    skill_path = SKILLS_DIR / "cursor-edit" / "SKILL.md"
    assert skill_path.exists(), f"cursor-edit SKILL.md not found: {skill_path}"


def test_cursor_edit_skill_valid():
    skill_path = SKILLS_DIR / "cursor-edit" / "SKILL.md"
    content = skill_path.read_text(encoding="utf-8")
    
    assert "name: cursor-edit" in content, "Missing name field"
    assert "description:" in content, "Missing description field"
    assert "---" in content, "Missing YAML frontmatter"


def test_cursor_edit_script_exists():
    script_path = SKILLS_DIR / "cursor-edit" / "scripts" / "apply.ps1"
    assert script_path.exists(), f"apply.ps1 not found: {script_path}"


def test_analyze_skill_exists():
    skill_path = SKILLS_DIR / "analyze" / "SKILL.md"
    assert skill_path.exists(), f"analyze SKILL.md not found: {skill_path}"


def test_verify_skill_exists():
    skill_path = SKILLS_DIR / "verify" / "SKILL.md"
    assert skill_path.exists(), f"verify SKILL.md not found: {skill_path}"


def test_all_skills_have_required_fields():
    required_fields = ["name:", "description:"]
    
    for skill_dir in SKILLS_DIR.iterdir():
        if skill_dir.is_dir():
            skill_md = skill_dir / "SKILL.md"
            if skill_md.exists():
                content = skill_md.read_text(encoding="utf-8")
                for field in required_fields:
                    assert field in content, f"{skill_dir.name}: Missing {field}"


if __name__ == "__main__":
    test_skills_directory_exists()
    test_cursor_edit_skill_exists()
    test_cursor_edit_skill_valid()
    test_cursor_edit_script_exists()
    test_analyze_skill_exists()
    test_verify_skill_exists()
    test_all_skills_have_required_fields()
    print("All skill tests passed!")
