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
    # 1. Fetch previous git tag to determine commit range
    prev_tag = run_cmd(["git", "describe", "--tags", "--abbrev=0", "HEAD^"])
    if not prev_tag:
        prev_tag = run_cmd(["git", "describe", "--tags", "--abbrev=0"])
    
    # 2. Get commit list since the last tag (fallback to last 15 commits if tag is missing)
    git_range = f"{prev_tag}..HEAD" if prev_tag else "HEAD~15..HEAD"
    log_format = "%s (%h)"
    commits_raw = run_cmd(["git", "log", f"--format={log_format}", "--no-merges", git_range])
    
    if not commits_raw:
        commits_raw = run_cmd(["git", "log", f"--format={log_format}", "--no-merges", "-n", "15"])
    
    commits = commits_raw.splitlines() if commits_raw else []
    
    # 3. Classify commits based on prefixes
    groups = {
        "fix": [],
        "feat": [],
        "perf": [],
        "chore": []
    }
    
    uncategorized = []
    
    for commit in commits:
        commit_lower = commit.lower()
        if commit_lower.startswith("fix:") or commit_lower.startswith("fix("):
            groups["fix"].append(commit)
        elif commit_lower.startswith("feat:") or commit_lower.startswith("feat("):
            groups["feat"].append(commit)
        elif commit_lower.startswith("perf:") or commit_lower.startswith("perf("):
            groups["perf"].append(commit)
        elif commit_lower.startswith("chore:") or commit_lower.startswith("chore("):
            groups["chore"].append(commit)
        elif any(commit_lower.startswith(prefix) for prefix in ["refactor:", "refactor(", "docs:", "docs(", "style:", "style(", "test:", "test(", "build:", "build(", "ci:", "ci("]):
            groups["chore"].append(commit)
        else:
            uncategorized.append(commit)
            
    # 4. Build changelog body
    changelog_parts = []
    has_conventional = any(len(v) > 0 for v in groups.values())
    
    if has_conventional:
        headers = {
            "feat": "✨ Features",
            "fix": "🐛 Fixes",
            "perf": "⚡ Performance",
            "chore": "🔧 Chores"
        }
        for key in ["feat", "fix", "perf", "chore"]:
            if groups[key]:
                changelog_parts.append(f"### {headers[key]}")
                for commit in groups[key]:
                    changelog_parts.append(f"- {commit}")
                changelog_parts.append("")
        
        if uncategorized:
            changelog_parts.append("### 📝 General Changes")
            for commit in uncategorized:
                changelog_parts.append(f"- {commit}")
            changelog_parts.append("")
    else:
        changelog_parts.append("### 📝 General Changes")
        for commit in commits:
            changelog_parts.append(f"- {commit}")
        changelog_parts.append("")
        
    changelog_markdown = "\n".join(changelog_parts).strip()
    
    # 5. Extract base kernel version from Makefile
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
    
    # Prepend release header block
    header = (
        "## 👑 Epitaph Kernel Release\n"
        "- **Device:** Redmi 12 / fire\n"
        "- **Kernel Version:** GKI 6.6\n"
        "- **Base Kernel Version:** `common-android15-6.6` (v" + base_kernel_ver + ")\n"
        "- **Build Date:** `" + build_date + "`\n\n"
        "---\n\n"
    )
    
    full_body = header + changelog_markdown
    
    with open("changelog.md", "w") as f:
        f.write(full_body)
        
    print("✅ Changelog generated successfully.")

if __name__ == "__main__":
    main()
