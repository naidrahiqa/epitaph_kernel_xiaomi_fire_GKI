import struct, re, gzip, io

boot_img = r"D:\fire_global_images_OS2.0.206.0.VMXMIXM_15.0\images\boot.img"

f = open(boot_img, "rb")
header = f.read(4096)

kernel_size = struct.unpack_from("<I", header, 8)[0]
print("Kernel size in boot.img: {} bytes ({:.1f} MB)".format(kernel_size, kernel_size / 1024 / 1024))

# Boot header v3: page size is always 4096
page_size = 4096
f.seek(page_size)
kernel_data = f.read(kernel_size)
f.close()

print("First 16 bytes of kernel: {}".format(kernel_data[:16].hex()))

# Check if kernel is gzip compressed
if kernel_data[:2] == b"\x1f\x8b":
    print("Kernel is GZIP compressed, decompressing...")
    try:
        decompressed = gzip.decompress(kernel_data)
        print("Decompressed size: {} bytes ({:.1f} MB)".format(len(decompressed), len(decompressed) / 1024 / 1024))
        kernel_data = decompressed
    except Exception as e:
        print("Gzip decompress error: {}".format(e))
        # Try partial decompression
        try:
            d = gzip.GzipFile(fileobj=io.BytesIO(kernel_data))
            kernel_data = d.read(8 * 1024 * 1024)
            print("Partial decompress: {} bytes".format(len(kernel_data)))
        except Exception as e2:
            print("Partial decompress also failed: {}".format(e2))
elif kernel_data[:4] == b"\x04\x22\x4d\x18":
    print("Kernel is LZ4 compressed")
elif kernel_data[:4] == b"\x28\xb5\x2f\xfd":
    print("Kernel is ZSTD compressed")
else:
    # Check for ARM64 kernel magic
    arm64_magic = struct.unpack_from("<I", kernel_data, 56)[0] if len(kernel_data) > 60 else 0
    if arm64_magic == 0x644d5241:
        print("Raw ARM64 kernel Image detected")
    else:
        print("Unknown kernel format, trying to find gzip header...")
        gz_offset = kernel_data.find(b"\x1f\x8b\x08")
        if gz_offset >= 0:
            print("Found gzip at offset {}".format(gz_offset))
            try:
                d = gzip.GzipFile(fileobj=io.BytesIO(kernel_data[gz_offset:]))
                kernel_data = d.read(8 * 1024 * 1024)
                print("Decompressed {} bytes from offset {}".format(len(kernel_data), gz_offset))
            except Exception as e:
                print("Decompress failed: {}".format(e))

# Search for version strings
print("\n=== Searching for kernel version info ===")

found = False
for match in re.finditer(b"Linux version (\\d+\\.\\d+\\.\\d+[^\\x00\\n]*)", kernel_data):
    ver = match.group(1).decode("ascii", errors="replace")
    print("KERNEL VERSION: {}".format(ver))
    found = True
    break

for match in re.finditer(b"(\\d+\\.\\d+\\.\\d+-android\\d+-\\d+)", kernel_data):
    ver = match.group(1).decode("ascii", errors="replace")
    print("KMI STRING: {}".format(ver))
    found = True
    break

for match in re.finditer(b"vermagic=(\\S+)", kernel_data):
    print("VERMAGIC: {}".format(match.group(1).decode("ascii", errors="replace")))
    found = True
    break

if not found:
    print("No version strings found in kernel data")
    # Dump any strings that look like version info
    for match in re.finditer(b"6\\.6\\.\\d+", kernel_data[:2*1024*1024]):
        ctx_start = max(0, match.start() - 20)
        ctx_end = min(len(kernel_data), match.end() + 40)
        ctx = kernel_data[ctx_start:ctx_end]
        printable = ctx.decode("ascii", errors="replace")
        print("  Found 6.6.x at offset {}: {}".format(match.start(), repr(printable)))

print("\nDone.")
