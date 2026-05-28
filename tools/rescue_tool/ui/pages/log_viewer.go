package pages

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
	"github.com/naidrahiqa/epitaph_rescue/internal/parser"
)



type LogViewerPage struct {
	deviceMgr    *adb.DeviceManager
	parser       *parser.Parser
	window       fyne.Window

	// Data
	logLines     []string
	logPath      string
	analysisRes  *parser.AnalysisResult

	// Stream control
	streamingActive bool
	stopStreamChan  chan struct{}
	logModeSelect   *widget.Select
	btnStream       *widget.Button

	// Widgets
	lblSummary   *widget.Label
	lblDiagnosis *widget.Label
	lstLog       *widget.List
	searchEntry  *widget.Entry
	filteredLines []int // indices into logLines

	// Action buttons
	btnPullLog   *widget.Button
	btnOpenFile  *widget.Button
	btnCopy      *widget.Button
	btnNotepad   *widget.Button

	mainContent  *fyne.Container
}

func NewLogViewerPage(dm *adb.DeviceManager, w fyne.Window) *LogViewerPage {
	lv := &LogViewerPage{
		deviceMgr: dm,
		parser:    parser.NewParser(),
		window:    w,
		logLines:  []string{"Belum ada log yang dimuat.", "Pilih Mode Log di atas untuk memulai streaming atau analisis crash log manual."},
	}
	lv.filterLines("")
	lv.buildUI()
	return lv
}

func (lv *LogViewerPage) buildUI() {
	// Header Info Panel
	lv.lblSummary = widget.NewLabel("Status: Tidak ada log")
	lv.lblSummary.TextStyle = fyne.TextStyle{Bold: true}

	lv.lblDiagnosis = widget.NewLabel("Silakan muat file crash log (last_kmsg / console-ramoops) untuk memulai analisis otomatis.")
	lv.lblDiagnosis.Wrapping = fyne.TextWrapWord

	diagnosisCard := NewNeoCard("Hasil Analisis Log", "Kemungkinan penyebab error", lv.lblDiagnosis)

	// Mode Selector & Stream buttons
	lv.btnPullLog = widget.NewButtonWithIcon("Pull dari Device", theme.DownloadIcon(), func() {
		lv.pullLogFromDevice()
	})
	lv.btnPullLog.Importance = widget.HighImportance

	lv.btnStream = widget.NewButtonWithIcon("Start Live Stream", theme.MediaPlayIcon(), func() {
		if lv.streamingActive {
			if lv.stopStreamChan != nil {
				close(lv.stopStreamChan)
			}
		} else {
			lv.startLiveStream(lv.logModeSelect.Selected)
		}
	})
	lv.btnStream.Importance = widget.HighImportance
	lv.btnStream.Hide() // Starts hidden because default is RAMoops

	lv.logModeSelect = widget.NewSelect([]string{"RAMoops (Crash Log)", "Live Logcat", "Live Dmesg"}, func(s string) {
		if s == "RAMoops (Crash Log)" {
			lv.btnStream.Hide()
			lv.btnPullLog.Show()
		} else {
			lv.btnStream.Show()
			lv.btnPullLog.Hide()
		}
	})
	lv.logModeSelect.SetSelected("RAMoops (Crash Log)")

	// Action Buttons
	lv.btnOpenFile = widget.NewButtonWithIcon("Buka File Log", theme.FolderOpenIcon(), func() {
		lv.openFileDialog()
	})

	lv.btnCopy = widget.NewButtonWithIcon("Copy Log", theme.ContentCopyIcon(), func() {
		lv.copyLogToClipboard()
	})
	lv.btnCopy.Disable()

	lv.btnNotepad = widget.NewButtonWithIcon("Buka di Notepad", theme.DocumentCreateIcon(), func() {
		lv.openInNotepad()
	})
	lv.btnNotepad.Disable()

	// Search & Filter
	lv.searchEntry = widget.NewEntry()
	lv.searchEntry.SetPlaceHolder("Cari kata kunci di log... (cth: Kernel panic, KSU, avc)")
	lv.searchEntry.OnChanged = func(q string) {
		lv.filterLines(q)
	}

	modeContainer := container.NewHBox(
		widget.NewLabel("Pilih Jenis Log:"),
		lv.logModeSelect,
	)

	btnContainer := container.NewAdaptiveGrid(4,
		container.NewMax(lv.btnPullLog, lv.btnStream),
		lv.btnOpenFile,
		lv.btnCopy,
		lv.btnNotepad,
	)

	// Log List View
	lv.lstLog = widget.NewList(
		func() int {
			return len(lv.filteredLines)
		},
		func() fyne.CanvasObject {
			lineNum := canvas.NewText("00000", colorTextMuted)
			lineNum.TextStyle = fyne.TextStyle{Monospace: true}
			
			lineText := canvas.NewText("", colorTextPri)
			lineText.TextStyle = fyne.TextStyle{Monospace: true}
			
			return container.NewHBox(lineNum, lineText)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(lv.filteredLines) {
				return
			}
			origIdx := lv.filteredLines[id]
			if origIdx < 0 || origIdx >= len(lv.logLines) {
				return
			}
			lineContent := lv.logLines[origIdx]

			hbox := item.(*fyne.Container)
			numText := hbox.Objects[0].(*canvas.Text)
			content := hbox.Objects[1].(*canvas.Text)

			numText.Text = fmt.Sprintf("%5d │ ", origIdx+1)

			dispText := lineContent
			if len(dispText) > 250 {
				dispText = dispText[:247] + "..."
			}
			content.Text = dispText

			hl := lv.parser.HighlightLine(lineContent)
			switch hl {
			case "critical":
				content.Color = colorError
			case "warning":
				content.Color = colorWarning
			case "ksu":
				content.Color = colorSecondary
			case "susfs":
				content.Color = colorSuccess
			default:
				content.Color = colorTextPri
			}
			
			numText.Refresh()
			content.Refresh()
		},
	)

	listWrapper := container.NewScroll(lv.lstLog)
	listWrapper.SetMinSize(fyne.NewSize(0, 200))

	// Settings Accordion
	lblPath := widget.NewLabel("Folder: " + GetLogOutputDir())
	lblPath.TextStyle = fyne.TextStyle{Monospace: true}
	lblPath.Wrapping = fyne.TextWrapWord

	btnChangeFolder := widget.NewButtonWithIcon("Ubah Folder", theme.SearchReplaceIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			path := uri.Path()
			SetLogOutputDir(path)
			lblPath.SetText("Folder: " + path)
			logger.Info("Log output directory updated to: %s", path)
		}, lv.window)
	})
	btnChangeFolder.Importance = widget.LowImportance

	btnOpenFolder := widget.NewButtonWithIcon("Buka Folder", theme.FolderOpenIcon(), func() {
		OpenFolderInExplorer(GetLogOutputDir())
	})
	btnOpenFolder.Importance = widget.LowImportance

	btnResetFolder := widget.NewButtonWithIcon("Reset Default", theme.ViewRefreshIcon(), func() {
		if fyne.CurrentApp() != nil {
			fyne.CurrentApp().Preferences().RemoveValue("crash_log_dir")
		}
		path := GetLogOutputDir()
		lblPath.SetText("Folder: " + path)
		logger.Info("Log output directory reset to default: %s", path)
	})
	btnResetFolder.Importance = widget.LowImportance

	settingsContent := container.NewVBox(
		lblPath,
		container.NewAdaptiveGrid(3,
			btnChangeFolder,
			btnOpenFolder,
			btnResetFolder,
		),
	)

	accordion := widget.NewAccordion(
		widget.NewAccordionItem("Pengaturan Folder Penyimpanan Log", settingsContent),
	)

	// Main Layout
	lv.mainContent = container.NewBorder(
		container.NewVBox(
			lv.lblSummary,
			diagnosisCard,
			modeContainer,
			btnContainer,
			accordion,
			widget.NewSeparator(),
			lv.searchEntry,
		),
		nil, // bottom
		nil, // left
		nil, // right
		listWrapper, // center
	)
}

func (lv *LogViewerPage) Content() fyne.CanvasObject {
	return lv.mainContent
}

func (lv *LogViewerPage) filterLines(query string) {
	query = strings.ToLower(query)
	var filtered []int
	for i, line := range lv.logLines {
		if query == "" || strings.Contains(strings.ToLower(line), query) {
			filtered = append(filtered, i)
		}
	}
	lv.filteredLines = filtered
	if lv.lstLog != nil {
		lv.lstLog.Refresh()
	}
}

func (lv *LogViewerPage) pullLogFromDevice() {
	info := lv.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di mode Android (ADB). Hidupkan HP dan aktifkan USB Debugging."), lv.window)
		return
	}

	lv.btnPullLog.Disable()
	lv.lblSummary.SetText("Status: Menarik log dari device...")
	lv.mainContent.Refresh()

	go func() {
		adbClient := lv.deviceMgr.GetADBClient()
		pstorePaths := []string{
			"/sys/fs/pstore/console-ramoops-0",
			"/sys/fs/pstore/dmesg-ramoops-0",
			"/sys/fs/pstore/console-ramoops",
			"/proc/last_kmsg",
		}

		var content string
		var matchedPath string

		for _, path := range pstorePaths {
			out, err := adbClient.Shell(fmt.Sprintf("cat %s 2>/dev/null", path))
			if err == nil && strings.TrimSpace(out) != "" && len(out) > 100 {
				content = out
				matchedPath = path
				break
			}
		}

		fyne.Do(func() {
			lv.btnPullLog.Enable()
			if content == "" {
				dialog.ShowError(fmt.Errorf("gagal menarik log dari PStore. Device bootloop sebelumnya mungkin tidak menggunakan kernel dengan PStore/RAMoops diaktifkan."), lv.window)
				lv.lblSummary.SetText("Status: Gagal menarik log")
				lv.mainContent.Refresh()
				return
			}

			outputDir := GetLogOutputDir()
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				logger.Error("Failed to create log directory: %v", err)
			}
			timestamp := time.Now().Format("20060102_150405")
			tmpFile := filepath.Join(outputDir, fmt.Sprintf("pulled_ramoops_%s.txt", timestamp))
			if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
				logger.Error("Failed to save log file: %v", err)
			}
			lv.logPath = tmpFile

			lv.loadLogContent(content)
			lv.lblSummary.SetText(fmt.Sprintf("Status: Berhasil menarik log dari %s", matchedPath))
			dialog.ShowInformation("Berhasil!", fmt.Sprintf("Log berhasil ditarik dan disimpan ke:\n%s", tmpFile), lv.window)
		})
	}()
}

func (lv *LogViewerPage) startLiveStream(mode string) {
	info := lv.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di mode Android (ADB). Hubungkan kabel USB dan aktifkan USB Debugging."), lv.window)
		return
	}

	lv.streamingActive = true
	lv.stopStreamChan = make(chan struct{})
	lv.btnStream.SetText("Stop Live Stream")
	lv.btnStream.SetIcon(theme.CancelIcon())
	lv.btnOpenFile.Disable()
	lv.logModeSelect.Disable()
	lv.lblSummary.SetText(fmt.Sprintf("Status: Live Streaming %s...", mode))

	go func() {
		adbPath := lv.deviceMgr.GetADBClient().BinaryPath
		var args []string
		if mode == "Live Logcat" {
			args = []string{"logcat", "-v", "time"}
		} else {
			// Live Dmesg (requires su for some Android versions, fallback to standard dmesg)
			args = []string{"shell", "su -c 'dmesg -w' 2>/dev/null || dmesg -w 2>/dev/null || dmesg"}
		}

		ctx, cancel := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, adbPath, args...)
		
		adb.PrepareCmd(cmd)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("gagal membuka log pipe: %v", err), lv.window)
			})
			return
		}

		if err := cmd.Start(); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("gagal memulai streaming process: %v", err), lv.window)
			})
			return
		}

		reader := bufio.NewReader(stdout)

		fyne.Do(func() {
			lv.logLines = []string{fmt.Sprintf("=== MULAI STREAMING %s ===", strings.ToUpper(mode))}
			lv.filterLines("")
		})

		// Wait in goroutine for stop trigger
		go func() {
			<-lv.stopStreamChan
			cancel()
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		}()

		var batchMu sync.Mutex
		var lineBatch []string
		flushTicker := time.NewTicker(100 * time.Millisecond)
		stopFlushChan := make(chan struct{})

		// Background batch flusher goroutine
		go func() {
			for {
				select {
				case <-flushTicker.C:
					batchMu.Lock()
					if len(lineBatch) == 0 {
						batchMu.Unlock()
						continue
					}
					batchToFlush := lineBatch
					lineBatch = nil
					batchMu.Unlock()

					fyne.Do(func() {
						if len(lv.logLines) > 1500 {
							overCount := len(lv.logLines) + len(batchToFlush) - 1500
							if overCount > 0 {
								if overCount >= len(lv.logLines) {
									lv.logLines = nil
								} else {
									lv.logLines = lv.logLines[overCount:]
								}
							}
						}
						lv.logLines = append(lv.logLines, batchToFlush...)
						lv.filterLines(lv.searchEntry.Text)
					})
				case <-stopFlushChan:
					flushTicker.Stop()
					return
				}
			}
		}()

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			batchMu.Lock()
			lineBatch = append(lineBatch, line)
			batchMu.Unlock()
		}

		close(stopFlushChan)
		_ = cmd.Wait()

		fyne.Do(func() {
			lv.streamingActive = false
			lv.btnStream.SetText("Start Live Stream")
			lv.btnStream.SetIcon(theme.MediaPlayIcon())
			lv.btnOpenFile.Enable()
			lv.logModeSelect.Enable()
			lv.lblSummary.SetText(fmt.Sprintf("Status: Streaming %s selesai", mode))
			lv.lblDiagnosis.SetText("✅ Live stream selesai.\nAnda dapat menyalin log atau menyimpannya.")
			lv.btnCopy.Enable()
		})
	}()
}

func (lv *LogViewerPage) openFileDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(fmt.Errorf("gagal membaca file: %w", err), lv.window)
			return
		}

		lv.logPath = reader.URI().Path()
		lv.loadLogContent(string(data))
		lv.lblSummary.SetText(fmt.Sprintf("Status: Berhasil memuat %s", reader.URI().Name()))
	}, lv.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".log", ""}))
	fd.Show()
}

func (lv *LogViewerPage) loadLogContent(content string) {
	lv.logLines = strings.Split(content, "\n")
	lv.filterLines(lv.searchEntry.Text)

	// Perform Analysis
	lv.analysisRes = lv.parser.ParseLines(lv.logLines)

	// Update Diagnosis Card
	diagnosis := ""
	if len(lv.analysisRes.TopIssues) > 0 {
		for _, issue := range lv.analysisRes.TopIssues {
			emoji := "ℹ️"
			if issue.Severity == parser.CRITICAL {
				emoji = "🔴"
			} else if issue.Severity == parser.WARNING {
				emoji = "🟡"
			}
			diagnosis += fmt.Sprintf("%s [%s] %s\n   💡 Saran: %s (Baris %d)\n\n", 
				emoji, issue.Category, issue.Diagnosis, issue.ActionHint, issue.FirstLine)
		}
	} else {
		diagnosis = "✅ Tidak terdeteksi masalah kritis dari signature patterns.\nKemungkinan kernel crash karena alasan lain atau log tidak lengkap."
	}

	lv.lblDiagnosis.SetText(diagnosis)
	lv.btnCopy.Enable()
	lv.btnNotepad.Enable()
	lv.mainContent.Refresh()
}

func (lv *LogViewerPage) copyLogToClipboard() {
	fullText := strings.Join(lv.logLines, "\n")
	lv.window.Clipboard().SetContent(fullText)
	dialog.ShowInformation("Disalin!", "Log berhasil disalin ke clipboard.", lv.window)
}

func (lv *LogViewerPage) openInNotepad() {
	if lv.logPath == "" {
		return
	}
	_ = exec.Command("notepad.exe", lv.logPath).Start()
}
