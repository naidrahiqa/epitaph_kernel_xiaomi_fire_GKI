package pages

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/tools"
)

type HomePage struct {
	deviceMgr    *adb.DeviceManager
	toolsMgr     *tools.Manager
	window       fyne.Window
	
	// Thread control
	stopChan     chan struct{}
	
	// Navigation callback
	onNavigate   func(index int)
	
	// GUI Widgets
	statusHeader *widget.Label
	statusCircle *canvas.Circle
	
	lblModel     *widget.Label
	lblCodename  *widget.Label
	lblAndroid   *widget.Label
	lblROM       *widget.Label
	lblPanel     *widget.Label
	lblKSU       *widget.Label
	lblSUSFS     *widget.Label
	lblKernel    *widget.Label

	// Health Widgets
	lblCPUTemp        *widget.Label
	lblBatteryTemp     *widget.Label
	lblBatteryHealth   *widget.Label
	lblBatteryCurrent  *widget.Label
	lblBatteryVoltage  *widget.Label
	lblRAMStats        *widget.Label

	// Quick Action buttons
	btnRebootNormal    *widget.Button
	btnRebootRecovery  *widget.Button
	btnRebootFastboot  *widget.Button
	btnSlotSwitch      *widget.Button

	// Tools widgets
	lblToolsStatus *widget.Label
	btnDownload    *widget.Button
	progressBar    *widget.ProgressBar
	
	mainBox        *fyne.Container
}

func NewHomePage(dm *adb.DeviceManager, tm *tools.Manager, w fyne.Window) *HomePage {
	hp := &HomePage{
		deviceMgr: dm,
		toolsMgr:  tm,
		window:    w,
		stopChan:  make(chan struct{}),
	}
	
	hp.buildUI()
	return hp
}

func (hp *HomePage) SetOnNavigate(onNavigate func(index int)) {
	hp.onNavigate = onNavigate
}

func (hp *HomePage) buildUI() {
	// Status indicator header
	hp.statusCircle = canvas.NewCircle(color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff})
	hp.statusCircle.Resize(fyne.NewSize(14, 14))

	hp.statusHeader = NewNeoHeading("Device Status: Disconnected")

	statusContainer := container.NewHBox(
		layout.NewSpacer(),
		hp.statusCircle,
		widget.NewLabel(" "),
		hp.statusHeader,
		layout.NewSpacer(),
	)

	// Welcome Banner
	lblWelcomeTitle := widget.NewLabel("Selamat Datang di Epitaph Rescue")
	lblWelcomeTitle.TextStyle = fyne.TextStyle{Bold: true}
	lblWelcomeDesc := NewNeoLabel("Asisten pemulihan instan, analisis crash log kernel, pemindaian kompatibilitas hardware, serta toolbox universal untuk HP Android Anda.")
	welcomeCard := NewNeoCard("", "", container.NewVBox(
		lblWelcomeTitle,
		lblWelcomeDesc,
	))

	// Device details card / grid
	hp.lblModel = widget.NewLabel("—")
	hp.lblCodename = widget.NewLabel("—")
	hp.lblAndroid = widget.NewLabel("—")
	hp.lblROM = widget.NewLabel("—")
	hp.lblPanel = widget.NewLabel("—")
	hp.lblKSU = widget.NewLabel("—")
	hp.lblSUSFS = widget.NewLabel("—")
	hp.lblKernel = widget.NewLabel("—")

	specs := container.NewVBox(
		makeSpecRow("Model", hp.lblModel),
		makeSpecRow("Codename", hp.lblCodename),
		makeSpecRow("Android Version", hp.lblAndroid),
		makeSpecRow("ROM Build ID", hp.lblROM),
		makeSpecRow("Panel Variant", hp.lblPanel),
		makeSpecRow("KernelSU", hp.lblKSU),
		makeSpecRow("SUSFS Support", hp.lblSUSFS),
		makeSpecRow("Kernel Version", hp.lblKernel),
	)

	infoFooter := widget.NewLabel("Info dasar dibaca secara aman via standard user-space ADB getprop tanpa root.")
	infoFooter.TextStyle = fyne.TextStyle{Italic: true}
	infoFooter.Wrapping = fyne.TextWrapWord

	infoCard := NewNeoCard("Device Specification", "Live device parameters", container.NewVBox(
		specs,
		infoFooter,
	))

	// Device Health Card
	hp.lblCPUTemp = widget.NewLabel("—")
	hp.lblBatteryTemp = widget.NewLabel("—")
	hp.lblBatteryHealth = widget.NewLabel("—")
	hp.lblBatteryCurrent = widget.NewLabel("—")
	hp.lblBatteryVoltage = widget.NewLabel("—")
	hp.lblRAMStats = widget.NewLabel("—")

	healthSpecs := container.NewVBox(
		makeSpecRow("CPU Temp", hp.lblCPUTemp),
		makeSpecRow("Battery Temp", hp.lblBatteryTemp),
		makeSpecRow("Battery Health", hp.lblBatteryHealth),
		makeSpecRow("Charging Current", hp.lblBatteryCurrent),
		makeSpecRow("Battery Voltage", hp.lblBatteryVoltage),
		makeSpecRow("RAM Usage", hp.lblRAMStats),
	)
	healthCard := NewNeoCard("Device Health & Thermal", "Real-time hardware parameters", healthSpecs)

	// Quick Action Console Card
	makeRebootBtn := func(label string, icon fyne.Resource, target string) *widget.Button {
		return widget.NewButtonWithIcon(label, icon, func() { hp.executeReboot(target) })
	}
	hp.btnRebootNormal = makeRebootBtn("Normal", theme.ConfirmIcon(), "")
	hp.btnRebootRecovery = makeRebootBtn("Recovery", theme.SettingsIcon(), "recovery")
	hp.btnRebootFastboot = makeRebootBtn("Fastboot", theme.WarningIcon(), "bootloader")

	hp.btnSlotSwitch = widget.NewButtonWithIcon("Toggle Active Slot (A/B)", theme.ViewRefreshIcon(), func() {
		hp.toggleActiveSlot()
	})
	hp.btnSlotSwitch.Importance = widget.WarningImportance

	rebootBox := container.NewGridWithColumns(3,
		hp.btnRebootNormal,
		hp.btnRebootRecovery,
		hp.btnRebootFastboot,
	)

	quickActionCard := NewNeoCard("Quick Action Console", "Universal device controllers", container.NewVBox(
		NewNeoHeading("Reboot Device To:"),
		rebootBox,
		NeoDivider(),
		hp.btnSlotSwitch,
	))

	// Platform tools management
	hp.lblToolsStatus = widget.NewLabel("Checking platform-tools status...")
	hp.progressBar = widget.NewProgressBar()
	hp.progressBar.Hide()

	hp.btnDownload = widget.NewButtonWithIcon("Download platform-tools", theme.DownloadIcon(), func() {
		hp.btnDownload.Disable()
		hp.progressBar.Show()
		hp.progressBar.SetValue(0)
		hp.mainBox.Refresh()

		hp.toolsMgr.StartDownload(func(prog float64, status string) {
			fyne.Do(func() {
				hp.progressBar.SetValue(prog)
				hp.lblToolsStatus.SetText(status)
			})
		}, func(err error) {
			fyne.Do(func() {
				hp.progressBar.Hide()
				hp.btnDownload.Enable()
				if err != nil {
					hp.lblToolsStatus.SetText(fmt.Sprintf("Error: %v", err))
				} else {
					hp.lblToolsStatus.SetText("platform-tools installed successfully!")
					hp.btnDownload.Hide()
				}
				hp.updateToolsStatus()
			})
		})
	})

	toolsCard := NewNeoCard("Platform Tools", "Required dependencies for ADB/Fastboot", container.NewVBox(
		NeoDivider(),
		hp.lblToolsStatus,
		hp.progressBar,
		hp.btnDownload,
	))

	// Organize Left Column and Right Column
	leftColumn := infoCard
	rightColumn := container.NewVBox(
		toolsCard,
		healthCard,
		quickActionCard,
	)

	topGrid := container.NewAdaptiveGrid(2,
		leftColumn,
		rightColumn,
	)

	// Action Cards Section
	// 1. Rescue Card
	lblDescRescue := widget.NewLabel("Wizard 5 langkah untuk HP bootloop/stuck logo setelah flash kernel. Mendeteksi status HP, memandu flashing stock boot.img, menarik crash log, dan memberikan diagnosis.")
	lblDescRescue.Wrapping = fyne.TextWrapWord
	btnRescue := widget.NewButtonWithIcon("Jalankan Rescue Wizard", theme.WarningIcon(), func() {
		if hp.onNavigate != nil {
			hp.onNavigate(1)
		}
	})
	btnRescue.Importance = widget.HighImportance
	cardRescue := NewNeoCard("Rescue Wizard (Bootloop)", "HP Stuck Logo / Bootloop?", container.NewVBox(
		lblDescRescue,
		btnRescue,
	))

	// 2. Diagnose Card
	lblDescWifi := widget.NewLabel("Pemindaian mendalam otomatis terhadap hardware, driver kernel, SELinux, AVB bootloader, hingga RAMoops crash dumps.")
	lblDescWifi.Wrapping = fyne.TextWrapWord
	btnWifi := widget.NewButtonWithIcon("Scan Kernel & System", theme.SearchIcon(), func() {
		if hp.onNavigate != nil {
			hp.onNavigate(3)
		}
	})
	btnWifi.Importance = widget.HighImportance
	cardWifi := NewNeoCard("System & Kernel Diagnostics", "Auto-Scan & Error Diagnostic", container.NewVBox(
		lblDescWifi,
		btnWifi,
	))

	// 3. Log Card
	lblDescLog := widget.NewLabel("Tarik log crash RAMoops/PStore atau buka file log manual. Parser membaca ribuan baris dalam sekejap, menyoroti error dengan warna khusus.")
	lblDescLog.Wrapping = fyne.TextWrapWord
	btnLog := widget.NewButtonWithIcon("Buka Log Viewer", theme.DocumentIcon(), func() {
		if hp.onNavigate != nil {
			hp.onNavigate(2)
		}
	})
	btnLog.Importance = widget.HighImportance
	cardLog := NewNeoCard("Log Analyzer", "Ambil & Analisis Crash Log", container.NewVBox(
		lblDescLog,
		btnLog,
	))

	// 4. Pre-Flash Check Card
	lblDescValidate := widget.NewLabel("Verifikasi kompatibilitas sebelum instalasi Custom Kernel. Cek vbmeta, panel layar, versi OS/ROM untuk meminimalkan resiko bootloop.")
	lblDescValidate.Wrapping = fyne.TextWrapWord
	btnValidate := widget.NewButtonWithIcon("Mulai Pre-Flash Check", theme.ConfirmIcon(), func() {
		if hp.onNavigate != nil {
			hp.onNavigate(4)
		}
	})
	btnValidate.Importance = widget.HighImportance
	cardValidate := NewNeoCard("Pre-Flash Checker", "Sebelum Flash Custom Kernel", container.NewVBox(
		lblDescValidate,
		btnValidate,
	))

	actionsGrid := container.NewAdaptiveGrid(2,
		cardRescue,
		cardWifi,
		cardLog,
		cardValidate,
	)

	// Main Layout
	hp.mainBox = container.NewVBox(
		welcomeCard,
		statusContainer,
		topGrid,
		NeoDivider(),
		NewNeoHeading("Pilih Alur Pemulihan"),
		actionsGrid,
	)

	hp.updateToolsStatus()
}

func (hp *HomePage) updateToolsStatus() {
	if hp.toolsMgr.IsAvailable() {
		hp.lblToolsStatus.SetText("✅ platform-tools (ADB & Fastboot) detected and ready.")
		hp.btnDownload.Hide()
	} else {
		hp.lblToolsStatus.SetText("⚠️ platform-tools not found. Auto-download is required to run rescue operations.")
		hp.btnDownload.Show()
	}
	hp.mainBox.Refresh()
}

func (hp *HomePage) Content() fyne.CanvasObject {
	return container.NewScroll(hp.mainBox)
}

func (hp *HomePage) StartPolling() {
	go func() {
		ticker := time.NewTicker(hp.deviceMgr.PollInterval())
		defer ticker.Stop()
		
		// Initial detection
		hp.pollDevice()

		for {
			select {
			case <-ticker.C:
				hp.pollDevice()
			case <-hp.stopChan:
				return
			}
		}
	}()
}

func (hp *HomePage) StopPolling() {
	close(hp.stopChan)
}

func (hp *HomePage) pollDevice() {
	info := hp.deviceMgr.Detect()

	var battHealth, battTemp, battVoltage, battCurrent, cpuTemp, ramStats string
	if info.Mode == adb.ModeAndroid {
		adbClient := hp.deviceMgr.GetADBClient()

		// Query battery stats
		battOut, _ := adbClient.Shell("dumpsys battery 2>/dev/null")
		lines := strings.Split(battOut, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "health: ") {
				hVal := strings.TrimPrefix(line, "health: ")
				switch hVal {
				case "2":
					battHealth = "Good"
				case "3":
					battHealth = "Overheat"
				case "4":
					battHealth = "Dead"
				case "5":
					battHealth = "Over Voltage"
				case "7":
					battHealth = "Cold"
				default:
					battHealth = "Unknown"
				}
			} else if strings.HasPrefix(line, "temp: ") {
				tVal := strings.TrimPrefix(line, "temp: ")
				tFloat := 0.0
				fmt.Sscanf(tVal, "%f", &tFloat)
				battTemp = fmt.Sprintf("%.1f °C", tFloat/10.0)
			} else if strings.HasPrefix(line, "voltage: ") {
				vVal := strings.TrimPrefix(line, "voltage: ")
				vFloat := 0.0
				fmt.Sscanf(vVal, "%f", &vFloat)
				battVoltage = fmt.Sprintf("%.3f V", vFloat/1000.0)
			} else if strings.HasPrefix(line, "status: ") {
				sVal := strings.TrimPrefix(line, "status: ")
				switch sVal {
				case "2":
					battCurrent = "Charging"
				case "3":
					battCurrent = "Discharging"
				case "4":
					battCurrent = "Not charging"
				case "5":
					battCurrent = "Full"
				default:
					battCurrent = "Discharging"
				}
			}
		}

		// Battery Current
		currVal, _ := adbClient.Shell("cat /sys/class/power_supply/battery/current_now || cat /sys/class/power_supply/battery/current_avg || echo ''")
		currVal = strings.TrimSpace(currVal)
		if currVal != "" && currVal != "0" {
			currFloat := 0.0
			fmt.Sscanf(currVal, "%f", &currFloat)
			if currFloat > 10000 || currFloat < -10000 {
				battCurrent = fmt.Sprintf("%s (%.1f mA)", battCurrent, currFloat/1000.0)
			} else if currFloat > 0 || currFloat < 0 {
				battCurrent = fmt.Sprintf("%s (%.1f mA)", battCurrent, currFloat)
			}
		}

		// CPU Temp
		cpuTempVal, _ := adbClient.Shell("cat /sys/class/thermal/thermal_zone0/temp || cat /sys/class/thermal/thermal_zone1/temp || cat /sys/class/thermal/thermal_zone2/temp || echo ''")
		cpuTempVal = strings.TrimSpace(cpuTempVal)
		if cpuTempVal != "" {
			cpuVal := 0.0
			fmt.Sscanf(cpuTempVal, "%f", &cpuVal)
			if cpuVal > 1000 {
				cpuTemp = fmt.Sprintf("%.1f °C", cpuVal/1000.0)
			} else if cpuVal > 0 {
				cpuTemp = fmt.Sprintf("%.1f °C", cpuVal)
			} else {
				cpuTemp = "—"
			}
		} else {
			cpuTemp = "—"
		}

		// RAM stats
		ramOut, _ := adbClient.Shell("cat /proc/meminfo 2>/dev/null || echo ''")
		var memTotal, memAvailable float64
		for _, line := range strings.Split(ramOut, "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %f", &memTotal)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %f", &memAvailable)
			}
		}
		if memTotal > 0 {
			memUsed := memTotal - memAvailable
			ramStats = fmt.Sprintf("%.1f GB / %.1f GB (%.1f%%)", memUsed/1024.0/1024.0, memTotal/1024.0/1024.0, (memUsed/memTotal)*100.0)
		} else {
			ramStats = "—"
		}
	}

	fyne.Do(func() {
		// Update UI widgets
		switch info.Mode {
		case adb.ModeAndroid:
			hp.statusCircle.FillColor = color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff} // Green
			hp.statusHeader.SetText("Device Status: 🟢 Connected (Android)")
			
			hp.lblModel.SetText(nonEmpty(info.Model))
			hp.lblCodename.SetText(nonEmpty(info.Codename))
			hp.lblAndroid.SetText(nonEmpty(info.AndroidVersion))
			hp.lblROM.SetText(nonEmpty(info.ROMVersion))
			hp.lblPanel.SetText(nonEmpty(info.PanelVariant))
			
			if info.KSUVersion != "" {
				hp.lblKSU.SetText("✅ Active (" + info.KSUVersion + ")")
			} else {
				hp.lblKSU.SetText("❌ Not Detected")
			}
			
			if info.SUSFSActive {
				hp.lblSUSFS.SetText("✅ Active")
			} else {
				hp.lblSUSFS.SetText("❌ Inactive")
			}
			
			hp.lblKernel.SetText(nonEmpty(info.KernelVersion))

			// Health widgets
			hp.lblCPUTemp.SetText(nonEmpty(cpuTemp))
			hp.lblBatteryTemp.SetText(nonEmpty(battTemp))
			hp.lblBatteryHealth.SetText(nonEmpty(battHealth))
			hp.lblBatteryCurrent.SetText(nonEmpty(battCurrent))
			hp.lblBatteryVoltage.SetText(nonEmpty(battVoltage))
			hp.lblRAMStats.SetText(nonEmpty(ramStats))

			hp.btnRebootNormal.Enable()
			hp.btnRebootRecovery.Enable()
			hp.btnRebootFastboot.Enable()
			hp.btnSlotSwitch.Enable()

		case adb.ModeFastboot:
			hp.statusCircle.FillColor = color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff} // Amber/Orange
			hp.statusHeader.SetText("Device Status: 🟡 Connected (Fastboot Mode)")
			
			hp.lblModel.SetText("Fastboot Mode")
			hp.lblCodename.SetText(nonEmpty(info.Codename))
			hp.lblAndroid.SetText("—")
			hp.lblROM.SetText("—")
			hp.lblPanel.SetText("—")
			hp.lblKSU.SetText("—")
			hp.lblSUSFS.SetText("—")
			hp.lblKernel.SetText("—")

			// Health widgets
			hp.lblCPUTemp.SetText("—")
			hp.lblBatteryTemp.SetText("—")
			hp.lblBatteryHealth.SetText("—")
			hp.lblBatteryCurrent.SetText("—")
			hp.lblBatteryVoltage.SetText("—")
			hp.lblRAMStats.SetText("—")

			hp.btnRebootNormal.Enable()
			hp.btnRebootRecovery.Enable()
			hp.btnRebootFastboot.Disable()
			hp.btnSlotSwitch.Enable()

		default:
			hp.statusCircle.FillColor = color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff} // Red
			hp.statusHeader.SetText("Device Status: 🔴 Disconnected")
			
			hp.lblModel.SetText("—")
			hp.lblCodename.SetText("—")
			hp.lblAndroid.SetText("—")
			hp.lblROM.SetText("—")
			hp.lblPanel.SetText("—")
			hp.lblKSU.SetText("—")
			hp.lblSUSFS.SetText("—")
			hp.lblKernel.SetText("—")

			// Health widgets
			hp.lblCPUTemp.SetText("—")
			hp.lblBatteryTemp.SetText("—")
			hp.lblBatteryHealth.SetText("—")
			hp.lblBatteryCurrent.SetText("—")
			hp.lblBatteryVoltage.SetText("—")
			hp.lblRAMStats.SetText("—")

			hp.btnRebootNormal.Disable()
			hp.btnRebootRecovery.Disable()
			hp.btnRebootFastboot.Disable()
			hp.btnSlotSwitch.Disable()
		}

		hp.statusCircle.Refresh()
		hp.statusHeader.Refresh()
		hp.mainBox.Refresh()
	})
}

func nonEmpty(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	return s
}

func (hp *HomePage) executeReboot(target string) {
	info := hp.deviceMgr.GetCurrent()
	if info.Mode == adb.ModeNone {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi."), hp.window)
		return
	}

	progress := dialog.NewCustomWithoutButtons("Mengirim perintah reboot...", widget.NewActivity(), hp.window)
	progress.Show()

	go func() {
		var err error
		if info.Mode == adb.ModeFastboot {
			fbClient := hp.deviceMgr.GetFastbootClient()
			if target == "" {
				_, err = fbClient.Run("reboot")
			} else {
				_, err = fbClient.Run("reboot", target)
			}
		} else if info.Mode == adb.ModeAndroid {
			adbClient := hp.deviceMgr.GetADBClient()
			if target == "" {
				_, err = adbClient.Run("reboot")
			} else {
				_, err = adbClient.Run("reboot", target)
			}
		}

		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("gagal reboot: %v", err), hp.window)
			} else {
				dialog.ShowInformation("Rebooting", "Perangkat sedang melakukan booting ulang...", hp.window)
			}
		})
	}()
}

func (hp *HomePage) toggleActiveSlot() {
	info := hp.deviceMgr.GetCurrent()
	if info.Mode == adb.ModeNone {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi. Silakan hubungkan HP via USB."), hp.window)
		return
	}

	progress := dialog.NewCustomWithoutButtons("Memproses pemindahan slot...", widget.NewActivity(), hp.window)
	progress.Show()

	go func() {
		var err error
		var result string

		if info.Mode == adb.ModeFastboot {
			fbClient := hp.deviceMgr.GetFastbootClient()
			slotOut, slotErr := fbClient.Run("getvar", "current-slot")
			if slotErr != nil {
				err = slotErr
			} else {
				currentSlot := "a"
				if strings.Contains(strings.ToLower(slotOut), "current-slot: b") || strings.Contains(strings.ToLower(slotOut), "slot: b") || strings.Contains(strings.ToLower(slotOut), " b ") || strings.Contains(strings.ToLower(slotOut), "b\n") || strings.Contains(strings.ToLower(slotOut), "b\r") {
					currentSlot = "b"
				}

				targetSlot := "b"
				if currentSlot == "b" {
					targetSlot = "a"
				}

				_, flashErr := fbClient.Run("set_active", targetSlot)
				if flashErr != nil {
					err = flashErr
				} else {
					result = fmt.Sprintf("Berhasil memindahkan slot aktif Fastboot dari slot %s ke slot %s!", strings.ToUpper(currentSlot), strings.ToUpper(targetSlot))
				}
			}
		} else if info.Mode == adb.ModeAndroid {
			adbClient := hp.deviceMgr.GetADBClient()
			slotSuffix, slotErr := adbClient.Shell("getprop ro.boot.slot_suffix 2>/dev/null")
			slotSuffix = strings.TrimSpace(slotSuffix)
			if slotErr != nil || slotSuffix == "" {
				err = fmt.Errorf("gagal membaca slot A/B (apakah HP Anda mendukung A/B partition?): %v", slotErr)
			} else {
				currentSlot := strings.TrimPrefix(slotSuffix, "_") // e.g. "a" or "b"
				targetSlot := "b"
				targetIndex := "1"
				if currentSlot == "b" {
					targetSlot = "a"
					targetIndex = "0"
				}

				cmd := fmt.Sprintf("su -c 'bootctl set-active-boot-slot %s'", targetIndex)
				out, rootErr := adbClient.Shell(cmd)
				if rootErr != nil || strings.Contains(out, "Permission denied") || strings.Contains(out, "not found") {
					err = fmt.Errorf("pemindahan slot via ADB memerlukan akses root (HP tidak di-root).\n\n💡 Solusi: Masuk ke mode Fastboot lalu gunakan fitur ini untuk switch slot tanpa perlu root!")
				} else {
					result = fmt.Sprintf("Berhasil memindahkan slot boot via ADB dari slot %s ke slot %s!", strings.ToUpper(currentSlot), strings.ToUpper(targetSlot))
				}
			}
		}

		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(err, hp.window)
			} else {
				dialog.ShowInformation("Sukses", result, hp.window)
			}
		})
	}()
}

func makeSpecRow(name string, value *widget.Label) fyne.CanvasObject {
	nameLbl := widget.NewLabel(name)
	nameLbl.TextStyle = fyne.TextStyle{Monospace: true}
	value.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewBorder(nil, nil, nameLbl, nil, container.NewHBox(layout.NewSpacer(), value))
}
