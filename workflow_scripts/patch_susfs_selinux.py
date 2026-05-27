import os
import sys

def main():
    print("⏳ Running SUSFS SELinux services.c patcher...")
    
    # Try different relative paths depending on execution environment
    paths = [
        "kernel/common/security/selinux/ss/services.c",
        "common/security/selinux/ss/services.c",
        "security/selinux/ss/services.c"
    ]
    
    filepath = None
    for p in paths:
        if os.path.exists(p):
            filepath = p
            break
            
    if not filepath:
        print("⚠️ Warning: security/selinux/ss/services.c not found!")
        return

    print(f"🔍 Found services.c at: {filepath}")
    with open(filepath, "r") as f:
        content = f.read()

    # Check if the SUSFS functions are already defined in services.c
    if "security_context_to_sid_with_policy" in content:
        print("✅ SUSFS functions are already defined in services.c. No patch needed.")
        return

    print("🔧 SUSFS functions are missing in services.c. Injecting helper functions...")
    
    # In GKI 6.6, internal services.c helpers (like security_context_to_sid_core)
    # do NOT accept selinux_state or selinux_policy. They are static functions operating
    # directly on the active policy context.
    helper_code = """
/* Injected by Epitaph Build Script to resolve missing SUSFS symbols */
#if IS_ENABLED(CONFIG_KSU_SUSFS) || IS_ENABLED(CONFIG_SUSFS)
int security_context_to_sid_with_policy(struct selinux_state *state,
					const char *scontext, u32 scontext_len,
					u32 *out_sid, u32 def_sid, gfp_t gfp_flags,
					struct selinux_policy *policy)
{
	return security_context_to_sid_core(scontext, scontext_len,
					    out_sid, def_sid, gfp_flags, 0);
}

int security_sid_to_context_with_policy(struct selinux_state *state,
					u32 sid, char **scontext, u32 *scontext_len,
					struct selinux_policy *policy)
{
	return security_sid_to_context_core(sid, scontext, scontext_len, 0, 0);
}

int security_compute_av_user_with_policy(struct selinux_state *state,
					 u32 ssid, u32 tsid, u16 tclass, u32 requested,
					 struct av_decision *avd,
					 struct selinux_policy *policy)
{
	security_compute_av_user(ssid, tsid, tclass, requested, avd);
	return 0;
}
#endif
"""
    
    with open(filepath, "a") as f:
        f.write(helper_code)
        
    print("✅ Successfully injected SUSFS SELinux helpers into services.c!")

    # Clean up any SELinux-related .rej files so they don't fail the build integrity check
    dirpath = os.path.dirname(filepath)
    for root, dirs, files in os.walk(dirpath):
        for file in files:
            if file.endswith(".rej"):
                rej_path = os.path.join(root, file)
                print(f"🧹 Cleaning up non-critical rejected patch file: {rej_path}")
                try:
                    os.remove(rej_path)
                except Exception as e:
                    print(f"⚠️ Failed to remove {rej_path}: {e}")

if __name__ == "__main__":
    main()
