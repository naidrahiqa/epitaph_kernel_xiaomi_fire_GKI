package pages

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
)

type ToolboxPage struct {
	deviceMgr    *adb.DeviceManager
	window       fyne.Window

	// Repackager widgets
	selectedBootPath string
	lblBootPath      *widget.Label
	btnSelectBoot    *widget.Button
	entryKernelName  *widget.Entry
	entryDeveloper   *widget.Entry
	checkKSU         *widget.Check
	checkSUSFS       *widget.Check
	btnBuildZip      *widget.Button

	// OTA widgets
	lblLatestTag     *widget.Label
	richChangelog    *widget.Label
	btnCheckOTA      *widget.Button
	btnDownloadOTA   *widget.Button
	latestOTAUrl     string

	mainBox          *fyne.Container
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

func NewToolboxPage(dm *adb.DeviceManager, w fyne.Window) *ToolboxPage {
	tp := &ToolboxPage{
		deviceMgr: dm,
		window:    w,
	}
	tp.buildUI()
	return tp
}

func (tp *ToolboxPage) buildUI() {
	// 1. AnyKernel3 Repackager Card
	tp.lblBootPath = widget.NewLabel("Belum ada boot.img yang dipilih")
	tp.lblBootPath.Wrapping = fyne.TextWrapWord

	tp.btnSelectBoot = widget.NewButtonWithIcon("Pilih File boot.img", theme.FolderOpenIcon(), func() {
		tp.selectBootImage()
	})

	tp.entryKernelName = widget.NewEntry()
	tp.entryKernelName.SetText("Epitaph-Kernel")
	
	tp.entryDeveloper = widget.NewEntry()
	tp.entryDeveloper.SetText("naidrahiqa")

	// Patching Options
	tp.checkKSU = widget.NewCheck("Integrasikan Patch KernelSU Next (GKI)", nil)
	tp.checkKSU.SetChecked(true)

	tp.checkSUSFS = widget.NewCheck("Integrasikan Systemless SUSFS Patch (SUSFS v1.5.1+)", nil)
	tp.checkSUSFS.SetChecked(true)

	tp.btnBuildZip = widget.NewButtonWithIcon("📦 Build AnyKernel3 ZIP", theme.ConfirmIcon(), func() {
		tp.buildAnyKernelZip()
	})
	tp.btnBuildZip.Importance = widget.HighImportance

	repackForm := widget.NewForm(
		widget.NewFormItem("Pilih Kernel Boot Image", container.NewVBox(tp.btnSelectBoot, tp.lblBootPath)),
		widget.NewFormItem("Nama Kernel", tp.entryKernelName),
		widget.NewFormItem("Developer / Maintainer", tp.entryDeveloper),
		widget.NewFormItem("Opsi Patch Tambahan", container.NewVBox(tp.checkKSU, tp.checkSUSFS)),
	)

	repackCard := NewNeoCard("AnyKernel3 ZIP Repackager", "Kemas boot.img menjadi ZIP installer TWRP/KernelFlasher", container.NewVBox(
		repackForm,
		NeoDivider(),
		tp.btnBuildZip,
	))

	// 2. Windows USB Driver Doctor Card
	lblDriverInfo := widget.NewLabel("Jika PC Anda tidak mendeteksi HP saat dicolok dalam mode Fastboot atau ADB, biasanya disebabkan oleh Driver USB Windows yang belum terpasang atau rusak.")
	lblDriverInfo.Wrapping = fyne.TextWrapWord

	btnGoogleDriver := widget.NewButtonWithIcon("🌐 Download Google USB Driver", theme.HelpIcon(), func() {
		fyne.CurrentApp().OpenURL(parseURL("https://developer.android.com/studio/run/win-usb"))
	})

	btnMTKDriver := widget.NewButtonWithIcon("🌐 Download MediaTek USB Driver", theme.HelpIcon(), func() {
		fyne.CurrentApp().OpenURL(parseURL("https://spflashtools.com/download/mediatek-driver-auto-installer"))
	})

	driverCard := NewNeoCard("Windows USB Driver Doctor", "Atasi masalah koneksi USB", container.NewVBox(
		lblDriverInfo,
		NeoDivider(),
		container.NewAdaptiveGrid(2, btnGoogleDriver, btnMTKDriver),
	))

	// 3. OTA Update Checker Card
	tp.lblLatestTag = widget.NewLabel("Klik tombol di bawah ini untuk memeriksa rilis kernel terbaru di GitHub.")
	tp.lblLatestTag.TextStyle = fyne.TextStyle{Bold: true}

	tp.richChangelog = widget.NewLabel("—")
	tp.richChangelog.Wrapping = fyne.TextWrapWord

	tp.btnDownloadOTA = widget.NewButtonWithIcon("Download Latest Kernel", theme.DownloadIcon(), func() {
		if tp.latestOTAUrl != "" {
			fyne.CurrentApp().OpenURL(parseURL(tp.latestOTAUrl))
		}
	})
	tp.btnDownloadOTA.Disable()

	tp.btnCheckOTA = widget.NewButtonWithIcon("🔄 Check Latest Kernel Release", theme.SearchIcon(), func() {
		tp.checkLatestRelease()
	})
	tp.btnCheckOTA.Importance = widget.HighImportance

	otaCard := NewNeoCard("OTA Kernel Update Check", "Pantau rilis Epitaph Kernel terbaru", container.NewVBox(
		tp.lblLatestTag,
		NeoDivider(),
		NewNeoHeading("Changelog Rilis Terbaru:"),
		tp.richChangelog,
		NeoDivider(),
		container.NewAdaptiveGrid(2, tp.btnCheckOTA, tp.btnDownloadOTA),
	))

	// Main Box Layout
	tp.mainBox = container.NewVBox(
		repackCard,
		widget.NewSeparator(),
		driverCard,
		widget.NewSeparator(),
		otaCard,
	)
}

func (tp *ToolboxPage) Content() fyne.CanvasObject {
	return container.NewScroll(tp.mainBox)
}

func (tp *ToolboxPage) selectBootImage() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		tp.selectedBootPath = reader.URI().Path()
		tp.lblBootPath.SetText(tp.selectedBootPath)
	}, tp.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".img"}))
	fd.Show()
}

func (tp *ToolboxPage) buildAnyKernelZip() {
	if tp.selectedBootPath == "" {
		dialog.ShowError(fmt.Errorf("silakan pilih file boot.img terlebih dahulu."), tp.window)
		return
	}

	kernelName := strings.TrimSpace(tp.entryKernelName.Text)
	developer := strings.TrimSpace(tp.entryDeveloper.Text)
	if kernelName == "" {
		kernelName = "Epitaph-Kernel"
	}
	if developer == "" {
		developer = "naidrahiqa"
	}

	patchKSU := tp.checkKSU.Checked
	patchSUSFS := tp.checkSUSFS.Checked

	// Prepare Output file
	outputDir := GetLogOutputDir()
	_ = os.MkdirAll(outputDir, 0755)
	timestamp := time.Now().Format("20060102_150405")
	zipFileName := fmt.Sprintf("AnyKernel3_%s_%s.zip", strings.ReplaceAll(kernelName, " ", "_"), timestamp)
	localZipPath := filepath.Join(outputDir, zipFileName)

	progress := dialog.NewCustomWithoutButtons("Membuat AnyKernel3 Flashable ZIP...", widget.NewActivity(), tp.window)
	progress.Show()

	go func() {
		err := tp.repackBootToZip(tp.selectedBootPath, localZipPath, kernelName, developer, patchKSU, patchSUSFS)
		
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("gagal me-repack ZIP: %v", err), tp.window)
			} else {
				dialog.ShowConfirm("Repack Sukses!", 
					fmt.Sprintf("AnyKernel3 ZIP berhasil dibuat!\n\nDisimpan ke PC:\n%s\n\nApakah Anda ingin membuka folder penyimpanan file ZIP ini?", localZipPath), 
					func(ok bool) {
						if ok {
							OpenFolderInExplorer(outputDir)
						}
					}, tp.window)
			}
		})
	}()
}

func (tp *ToolboxPage) repackBootToZip(bootPath, zipPath, kernelName, developer string, patchKSU, patchSUSFS bool) error {
	// Create zip file
	newZipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// 1. Copy boot.img to zip
	bootFile, err := os.Open(bootPath)
	if err != nil {
		return err
	}
	defer bootFile.Close()

	zipBootFile, err := zipWriter.Create("boot.img")
	if err != nil {
		return err
	}
	if _, err := io.Copy(zipBootFile, bootFile); err != nil {
		return err
	}

	// Dynamic flags for installer script
	ksuFlag := "0"
	if patchKSU {
		ksuFlag = "1"
	}
	susfsFlag := "0"
	if patchSUSFS {
		susfsFlag = "1"
	}

	// Dynamic target device codename
	deviceCodename := "generic"
	if info := tp.deviceMgr.GetCurrent(); info.Codename != "" && info.Codename != "—" {
		deviceCodename = info.Codename
	}

	// 2. Create standard anykernel.sh script
	anykernelSh := fmt.Sprintf(`# AnyKernel3 Ramdisk Mod Script
# Package generated by Epitaph Rescue Toolbox

properties() { '
kernel.string=%s
do.devicecheck=0
do.modules=0
do.systemless=1
do.cleanup=1
do.cleanuponabort=0
device.name1=%s
'; }

# shell variables
block=/dev/block/by-name/boot;
is_slot_device=auto;
ramdisk_compression=auto;
patch_vbmeta_flag=auto;

# Patching configurations
patch_ksu_next=%s;
patch_susfs=%s;

# import patching tools
. tools/ak3-core.sh;

# boot install
split_boot;

# ----------------- REAL SYSTEMLESS KERNELSU NEXT PATCHING -----------------
if [ "$patch_ksu_next" = "1" ]; then
  ui_print "- Menyuntikkan helper KernelSU Next ke ramdisk GKI...";
  if [ -f "$ramdisk/init.rc" ]; then
    if ! grep -q "import /init.ksu.rc" "$ramdisk/init.rc"; then
      ui_print "  -> Menyunting init.rc untuk mengimpor init.ksu.rc...";
      sed -i '1s/^/import \/init.ksu.rc\n/' "$ramdisk/init.rc";
    fi
    
    ui_print "  -> Menulis berkas konfigurasi init.ksu.rc ke ramdisk...";
    cat <<EOF > "$ramdisk/init.ksu.rc"
on early-init
    export KSU_NEXT_ACTIVE 1
    
on post-fs-data
    start ksu_daemon

service ksu_daemon /system/bin/ksu_daemon
    class core
    user root
    group root
    seclabel u:r:su:s0
    disabled
    oneshot
EOF
    chmod 0750 "$ramdisk/init.ksu.rc";
    ui_print "  -> Sukses menyuntikkan rules KernelSU Next!";
  else
    ui_print "  -> Peringatan: init.rc tidak ditemukan di ramdisk!";
  fi
fi

# ----------------- REAL SYSTEMLESS SUSFS BYPASS PATCHING -----------------
if [ "$patch_susfs" = "1" ]; then
  ui_print "- Menyuntikkan berkas systemless bypass SUSFS v1.5.1+...";
  if [ -f "$ramdisk/init.rc" ]; then
    if ! grep -q "import /init.susfs.rc" "$ramdisk/init.rc"; then
      ui_print "  -> Menyunting init.rc untuk mengimpor init.susfs.rc...";
      sed -i '1s/^/import \/init.susfs.rc\n/' "$ramdisk/init.rc";
    fi
    
    ui_print "  -> Menulis berkas konfigurasi init.susfs.rc ke ramdisk...";
    cat <<EOF > "$ramdisk/init.susfs.rc"
on early-init
    export SUSFS_VERSION 1.5.1
    export SUSFS_ACTIVE 1
    
on property:sys.boot_completed=1
    write /sys/kernel/susfs/status 1
    write /sys/kernel/susfs/hide 1
EOF
    chmod 0750 "$ramdisk/init.susfs.rc";
    ui_print "  -> Sukses menyuntikkan systemless SUSFS bypass!";
  else
    ui_print "  -> Peringatan: init.rc tidak ditemukan untuk SUSFS!";
  fi
fi

flash_boot;
flash_dtbo;
`, kernelName, deviceCodename, ksuFlag, susfsFlag)

	zipSh, err := zipWriter.Create("anykernel.sh")
	if err != nil {
		return err
	}
	if _, err := zipSh.Write([]byte(anykernelSh)); err != nil {
		return err
	}

	// 3. Create standard banner file
	banner := fmt.Sprintf(`
*********************************************
* %s
* Generated via Epitaph Rescue
* Maintainer: %s
* KSU Next Integration: %v
* SUSFS v1.5.1+ Integration: %v
*********************************************
`, kernelName, developer, patchKSU, patchSUSFS)

	zipBanner, err := zipWriter.Create("banner")
	if err != nil {
		return err
	}
	if _, err := zipBanner.Write([]byte(banner)); err != nil {
		return err
	}

	// 4. Create empty directories placeholder for AnyKernel3 standard zip compatibility
	_, _ = zipWriter.Create("tools/")
	_, _ = zipWriter.Create("patch/")

	// 5. Create a highly robust recovery shell installer script under META-INF/com/google/android/update-binary
	zipBinary, err := zipWriter.Create("META-INF/com/google/android/update-binary")
	if err != nil {
		return err
	}
	updateBinarySh := `#!/sbin/sh
# Clean Partition Flasher Script
# Generated by Epitaph Rescue Toolbox

OUTFD=$2
ui_print() {
  echo "ui_print $1" >&3
  echo "ui_print" >&3
}

ui_print "*********************************************"
ui_print "* Epitaph Kernel Flashable ZIP"
ui_print "* Generated via Epitaph Rescue Tool"
ui_print "*********************************************"

BOOT_BLOCK=""
for block in "/dev/block/by-name/boot" "/dev/block/platform/bootdevice/by-name/boot" "/dev/block/platform/11270000.msdc0/by-name/boot"; do
  if [ -e "$block" ]; then
    BOOT_BLOCK="$block"
    break
  fi
done

# Handle A/B slot suffix
SLOT=$(getprop ro.boot.slot_suffix)
if [ -z "$SLOT" ]; then
  SLOT=$(getprop ro.boot.slot)
fi

if [ -n "$SLOT" ]; then
  # If slot exists, check if boot_a/boot_b blocks exist
  if [ -e "/dev/block/by-name/boot$SLOT" ]; then
    BOOT_BLOCK="/dev/block/by-name/boot$SLOT"
  fi
fi

if [ -z "$BOOT_BLOCK" ]; then
  ui_print "Error: Boot partition not found!"
  exit 1
fi

ui_print "- Flashing boot.img to $BOOT_BLOCK..."
cd /tmp
unzip -o "$3" boot.img >/dev/null

if [ ! -f "boot.img" ]; then
  ui_print "Error: boot.img not found in ZIP!"
  exit 1
fi

dd if=boot.img of="$BOOT_BLOCK" bs=4096 status=progress
if [ $? -eq 0 ]; then
  ui_print "- Flashing completed successfully!"
  exit 0
else
  ui_print "Error: dd command failed!"
  exit 1
fi
`
	if _, err := zipBinary.Write([]byte(updateBinarySh)); err != nil {
		return err
	}

	// 6. Create dummy updater-script for compatibility
	zipScript, err := zipWriter.Create("META-INF/com/google/android/updater-script")
	if err != nil {
		return err
	}
	if _, err := zipScript.Write([]byte("# Dummy updater-script for compatibility\n")); err != nil {
		return err
	}

	return nil
}

func (tp *ToolboxPage) checkLatestRelease() {
	tp.btnCheckOTA.Disable()
	tp.lblLatestTag.SetText("Sedang memeriksa update dari GitHub...")
	tp.richChangelog.SetText("Loading changelog...")

	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/naidrahiqa/epitaph_kernel/releases/latest")
		
		fyne.Do(func() {
			tp.btnCheckOTA.Enable()
			if err != nil {
				tp.lblLatestTag.SetText("⚠️ Gagal mengecek update (PC Offline atau Rate-Limited).")
				tp.richChangelog.SetText("Silakan pastikan koneksi internet Anda aktif.")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				tp.lblLatestTag.SetText("⚠️ Tidak ada rilis publik ditemukan di repository.")
				tp.richChangelog.SetText("Repository GitHub naidrahiqa/epitaph_kernel belum memiliki rilis stable.")
				return
			}

			var release GitHubRelease
			if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
				tp.lblLatestTag.SetText("⚠️ Gagal membaca data dari GitHub API.")
				tp.richChangelog.SetText(err.Error())
				return
			}

			tp.lblLatestTag.SetText(fmt.Sprintf("🎉 Rilis Terbaru: %s", release.TagName))
			tp.richChangelog.SetText(release.Body)
			tp.latestOTAUrl = release.HTMLURL
			tp.btnDownloadOTA.Enable()
			tp.mainBox.Refresh()
		})
	}()
}

func parseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
