import os
import re

def patch_function_body(content, func_pattern, replacement_body, func_name):
    """Replace the body of a function matching func_pattern with replacement_body.
    Returns (new_content, was_patched)."""
    match = re.search(func_pattern, content)
    if not match:
        return content, False

    start_idx = match.end()
    brace_count = 1
    end_idx = start_idx

    while brace_count > 0 and end_idx < len(content):
        char = content[end_idx]
        if char == '{':
            brace_count += 1
        elif char == '}':
            brace_count -= 1
        end_idx += 1

    if brace_count == 0:
        # Insert replacement_body right after the opening brace '{' (start_idx)
        # This keeps the original code intact as 'dead code', satisfying -Wunused-function checks!
        content = content[:start_idx] + replacement_body + content[start_idx:]
        print(f"  ✅ Patched {func_name} (via insertion)")
        return content, True

    return content, False


def main():
    print("⏳ Running Comprehensive Vermagic & Module Version Bypass Patch...")

    # All potential files where vermagic/modversion checks exist in GKI 6.6
    files_to_check = [
        "kernel/module/internal.h",
        "kernel/module/main.c",
        "kernel/module/version.c",
        "kernel/module.c",
        "common/kernel/module/internal.h",
        "common/kernel/module/main.c",
        "common/kernel/module/version.c",
        "common/kernel/module.c"
    ]

    patches_applied = 0

    for filepath in files_to_check:
        if not os.path.exists(filepath):
            continue

        print(f"🔍 Scanning: {filepath}")
        with open(filepath, "r") as f:
            content = f.read()
        
        original = content
        file_patched = False

        # PATCH 1: same_magic() — always return 1 (versions match)
        # This is the primary vermagic string comparison check
        content, p = patch_function_body(
            content,
            r"(?:static\s+inline\s+)?(?:int|bool)\s+same_magic\s*\([^)]*\)\s*\{",
            "\n\treturn 1;\n",
            "same_magic()"
        )
        file_patched = file_patched or p

        # PATCH 2: check_modinfo() — always return 0 (no error)
        # Secondary check that validates vermagic string via modinfo section
        content, p = patch_function_body(
            content,
            r"(?:static\s+)?int\s+check_modinfo\s*\([^)]*\)\s*\{",
            "\n\treturn 0;\n",
            "check_modinfo()"
        )
        file_patched = file_patched or p

        # PATCH 3: check_version() — always return true/1
        # CRC symbol version check (CONFIG_MODVERSIONS)
        # This prevents "disagrees about version of symbol" errors
        content, p = patch_function_body(
            content,
            r"(?:static\s+)?(?:int|bool)\s+check_version\s*\([^)]*\)\s*\{",
            "\n\treturn 1;\n",
            "check_version()"
        )
        file_patched = file_patched or p

        # PATCH 4: check_modstruct_version() — always return true/1
        # Module struct version check
        content, p = patch_function_body(
            content,
            r"(?:static\s+)?(?:int|bool)\s+check_modstruct_version\s*\([^)]*\)\s*\{",
            "\n\treturn 1;\n",
            "check_modstruct_version()"
        )
        file_patched = file_patched or p

        if file_patched:
            with open(filepath, "w") as f:
                f.write(content)
            patches_applied += 1
            print(f"✅ Saved patches to {filepath}")

    if patches_applied > 0:
        print(f"\n✅ Vermagic/modversion bypass applied to {patches_applied} file(s)")
    else:
        print("\n⚠️ WARNING: No vermagic/modversion functions found to patch!")
        print("   Stock Xiaomi vendor modules (wlan_drv_gen4m.ko) may fail to load!")

if __name__ == "__main__":
    main()
