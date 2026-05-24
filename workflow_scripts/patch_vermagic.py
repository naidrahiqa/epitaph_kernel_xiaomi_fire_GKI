import os
import re

def main():
    print("⏳ Running Vermagic Bypass Patch...")
    
    # Berkas potensial tempat fungsi same_magic berada di kernel Linux 6.6 / 6.1 / 5.15
    files_to_check = [
        "kernel/module/internal.h",
        "kernel/module/main.c",
        "kernel/module.c",
        "common/kernel/module/internal.h",
        "common/kernel/module/main.c",
        "common/kernel/module.c"
    ]
    
    patched = False
    
    for filepath in files_to_check:
        if os.path.exists(filepath):
            print(f"🔍 Menemukan berkas target: {filepath}")
            with open(filepath, "r") as f:
                content = f.read()
                
            # Regex untuk mencari deklarasi 'static inline int same_magic' beserta isi kurung kurawalnya
            # yang terkadang mencakup multiline
            match = re.search(r"static inline int same_magic\s*\([^)]*\)\s*\{", content)
            if match:
                print(f"🎯 Menemukan fungsi same_magic pada {filepath}!")
                
                # Kita cari kurung kurawal tutup pasangan dari '{' tersebut
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
                    # Kita ganti seluruh isi tubuh fungsi same_magic dengan "return 1;" secara absolut!
                    new_body = "\n\treturn 1;\n"
                    content = content[:start_idx] + new_body + content[end_idx-1:]
                    
                    with open(filepath, "w") as f:
                        f.write(content)
                    print(f"✅ Berhasil mem-patch same_magic pada {filepath}!")
                    patched = True
                    break
                    
    if not patched:
        print("⚠️ Peringatan: Tidak dapat menemukan fungsi same_magic untuk di-patch.")

if __name__ == "__main__":
    main()
