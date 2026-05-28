package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

const (
	WindowsSDKURL = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
)

type Manager struct {
	mu           sync.Mutex
	adbPath      string
	fastbootPath string
	downloading  bool
	progress     float64
	statusText   string
}

func NewManager() *Manager {
	m := &Manager{
		statusText: "Ready",
	}
	m.resolvePaths()
	return m
}

func (m *Manager) resolvePaths() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check if adb/fastboot are in system PATH
	if p, err := exec.LookPath("adb"); err == nil {
		m.adbPath = p
	}
	if p, err := exec.LookPath("fastboot"); err == nil {
		m.fastbootPath = p
	}

	// 2. If not found in PATH, check our local %APPDATA%/KernelRescueTool/platform-tools directory
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	localDir := filepath.Join(appData, "KernelRescueTool", "platform-tools")
	
	localADB := filepath.Join(localDir, "adb.exe")
	localFastboot := filepath.Join(localDir, "fastboot.exe")

	if m.adbPath == "" {
		if _, err := os.Stat(localADB); err == nil {
			m.adbPath = localADB
		}
	}
	if m.fastbootPath == "" {
		if _, err := os.Stat(localFastboot); err == nil {
			m.fastbootPath = localFastboot
		}
	}

	// 3. Fallback: if still empty, set expected local path so they can be downloaded
	if m.adbPath == "" {
		m.adbPath = localADB
	}
	if m.fastbootPath == "" {
		m.fastbootPath = localFastboot
	}
}

func (m *Manager) GetPaths() (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.adbPath, m.fastbootPath
}

func (m *Manager) IsAvailable() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	_, errA := os.Stat(m.adbPath)
	_, errF := os.Stat(m.fastbootPath)
	return errA == nil && errF == nil
}

func (m *Manager) GetStatus() (bool, float64, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.downloading, m.progress, m.statusText
}

func (m *Manager) StartDownload(onProgress func(float64, string), onComplete func(error)) {
	m.mu.Lock()
	if m.downloading {
		m.mu.Unlock()
		return
	}
	m.downloading = true
	m.progress = 0.0
	m.statusText = "Downloading platform-tools..."
	m.mu.Unlock()

	go func() {
		err := m.downloadAndExtract(onProgress)
		m.mu.Lock()
		m.downloading = false
		if err == nil {
			m.statusText = "Ready"
			m.progress = 1.0
			m.mu.Unlock()
			m.resolvePaths()
			onComplete(nil)
		} else {
			m.statusText = fmt.Sprintf("Download failed: %v", err)
			m.progress = 0.0
			m.mu.Unlock()
			onComplete(err)
		}
	}()
}

func (m *Manager) downloadAndExtract(onProgress func(float64, string)) error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	destDir := filepath.Join(appData, "KernelRescueTool")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	zipPath := filepath.Join(destDir, "platform-tools.zip")
	logger.Info("Downloading platform-tools from %s to %s", WindowsSDKURL, zipPath)

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Get(WindowsSDKURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	totalSize := resp.ContentLength
	var downloaded int64

	buffer := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buffer)
		if n > 0 {
			if _, werr := out.Write(buffer[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			if totalSize > 0 {
				prog := float64(downloaded) / float64(totalSize)
				m.mu.Lock()
				m.progress = prog * 0.7 // Save 30% for extraction
				m.statusText = fmt.Sprintf("Downloading... %.0f%%", prog*100)
				m.mu.Unlock()
				if onProgress != nil {
					onProgress(m.progress, m.statusText)
				}
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	out.Close()

	logger.Info("Extracting zip: %s to %s", zipPath, destDir)
	if err := unzip(zipPath, destDir, onProgress); err != nil {
		return err
	}

	// Delete temporary zip
	os.Remove(zipPath)

	logger.Info("platform-tools successfully extracted and verified")
	return nil
}

func unzip(src string, dest string, onProgress func(float64, string)) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	totalFiles := len(r.File)

	for idx, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		
		// Update progress visually during extraction
		if onProgress != nil && totalFiles > 0 {
			prog := 0.7 + (float64(idx)/float64(totalFiles))*0.3
			statusText := fmt.Sprintf("Extracting: %s... (%.0f%%)", filepath.Base(f.Name), prog*100)
			onProgress(prog, statusText)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
