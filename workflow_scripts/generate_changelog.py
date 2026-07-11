import subprocess
import os
from datetime import datetime

def run_cmd(cmd):
    """Run a shell command safely and return its trimmed output."""
    try:
        return subprocess.check_output(cmd, text=True, stderr=subprocess.DEVNULL).strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return ""

def main():
    # 1. Read curated CHANGELOG.md
    changelog_file = "CHANGELOG.md"
    changelog_body = ""
    if os.path.exists(changelog_file):
        with open(changelog_file, "r") as f:
            content = f.read().strip()
            # Find first version section
            import re
            match = re.search(r"## \[v\d+\].*?^(?=## \[v\d+\]|\Z)", content, re.MULTILINE | re.DOTALL)
            if match:
                changelog_body = match.group(0).strip()
    
    # 2. If no curated changelog found, fallback to git log (summary style)
    if not changelog_body:
        prev_tag = run_cmd(["git", "describe", "--tags", "--abbrev=0", "HEAD^"])
        if not prev_tag:
            prev_tag = run_cmd(["git", "describe", "--tags", "--abbrev=0"])
        git_range = f"{prev_tag}..HEAD" if prev_tag else "HEAD~15..HEAD"
        log_format = "%s"
        commits_raw = run_cmd(["git", "log", f"--format={log_format}", "--no-merges", git_range])
        if not commits_raw:
            commits_raw = run_cmd(["git", "log", f"--format={log_format}", "--no-merges", "-n", "15"])
        commits = commits_raw.splitlines() if commits_raw else []
        # Build flat list
        emoji_map = {
            "feat": "✨", "fix": "🐛", "perf": "⚡",
            "chore": "🔧", "docs": "📖", "ci": "👷",
            "refactor": "♻️", "test": "🧪", "style": "🎨"
        }
        changelog_parts = []
        for commit in commits:
            prefix = commit.split(":")[0].split("(")[0].lower()
            emoji = emoji_map.get(prefix, "📝")
            changelog_parts.append(f"- {emoji} {commit}")
        changelog_body = "\n".join(changelog_parts)
    
    # 3. Extract base kernel version from Makefile
    base_kernel_ver = "6.6"
    makefile_path = "kernel/common/Makefile"
    if os.path.exists(makefile_path):
        version = ""
        patchlevel = ""
        sublevel = ""
        with open(makefile_path, "r") as f:
            for line in f:
                if line.startswith("VERSION ="):
                    version = line.split("=")[1].strip()
                elif line.startswith("PATCHLEVEL ="):
                    patchlevel = line.split("=")[1].strip()
                elif line.startswith("SUBLEVEL ="):
                    sublevel = line.split("=")[1].strip()
                    break
        if version and patchlevel:
            base_kernel_ver = f"{version}.{patchlevel}"
            if sublevel:
                base_kernel_ver += f".{sublevel}"
    
    build_date = datetime.now().strftime("%Y-%m-%d")
    
    # 4. Build compact header
    header = (
        "**Epitaph Kernel** | Redmi 12 / fire | GKI 6.6 ("
        + base_kernel_ver + ") | " + build_date + "\n\n"
    )
    
    full_body = header + changelog_body + "\n"
    
    with open("changelog.md", "w") as f:
        f.write(full_body)
    
    print("✅ Changelog generated successfully.")

if __name__ == "__main__":
    main()
