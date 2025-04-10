package handle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"
	"golang.org/x/sys/windows/registry"
)

type TabPage interface {
	ID() string
	Title() string
	URL() string
	Domain() string
	IsClosed() bool
	BringToFront()
	OpenInNewTab(id string, action func() error, timeout float64) TabPage
	WaitSelector(selector string, timeout float64) playwright.Locator
	QuerySelector(selector string) playwright.Locator
	QuerySelectorAll(selector string) []playwright.Locator
	ClearLocalData() error
	Goto(url string) error
	Evaluate(expression string, arg ...any) (any, error)
	Page() playwright.Page
}

type Browser interface {
	Name() string
	Port() int
	TabPages() []TabPage
	NewTabPage(id string, url string) TabPage
	FindTabPage(id string) TabPage
	SwitchToTabPage(id string) error
	CloseTabPage(id string) error
	IsAlive() bool
	Close() error
}

// Windows实现
func findWindowsBrowsers(browsers map[string]string) {
	// 常见浏览器注册表路径
	browserPaths := map[string]string{
		"chrome":  `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe`,
		"edge":    `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\msedge.exe`,
		"firefox": `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\firefox.exe`,
		"opera":   `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\opera.exe`,
	}

	for name, regPath := range browserPaths {
		key, err := registry.OpenKey(
			registry.LOCAL_MACHINE,
			regPath,
			registry.QUERY_VALUE|registry.WOW64_64KEY,
		)
		if err != nil {
			continue
		}
		defer key.Close()

		path, _, err := key.GetStringValue("")
		if err != nil {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			browsers[name] = path
		}
	}
}

// macOS实现
func findMacBrowsers(browsers map[string]string) {
	appDirs := []string{
		"/Applications",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	apps := map[string]string{
		"chrome":  "Google Chrome.app",
		"firefox": "Firefox.app",
		"safari":  "Safari.app",
		"edge":    "Microsoft Edge.app",
		"opera":   "Opera.app",
	}

	for _, dir := range appDirs {
		for name, app := range apps {
			appPath := filepath.Join(dir, app)
			exePath := filepath.Join(appPath, "Contents/MacOS", strings.TrimSuffix(app, ".app"))

			if _, err := os.Stat(exePath); err == nil {
				browsers[name] = exePath
			}
		}
	}
}

// Linux实现
func findLinuxBrowsers(browsers map[string]string) {
	// 通过which命令查找
	common := []string{
		"google-chrome", "chrome",
		"chromium", "chromium-browser",
		"firefox", "microsoft-edge",
		"opera",
	}

	for _, exe := range common {
		path, err := exec.LookPath(exe)
		if err == nil {
			browsers[exe] = path
		}
	}

	// 检查flatpak应用
	flatpakApps := []string{
		"com.google.Chrome",
		"org.mozilla.firefox",
		"com.microsoft.Edge",
	}

	flatpakAppNames := map[string]string{
		"com.google.Chrome":   "chrome",
		"org.mozilla.firefox": "firefox",
		"com.microsoft.Edge":  "edge",
	}

	for _, app := range flatpakApps {
		cmd := exec.Command("flatpak", "info", "--show-location", app)
		if path, err := cmd.Output(); err == nil {
			exePath := filepath.Join(
				strings.TrimSpace(string(path)),
				"files", "bin", app,
			)
			if _, err := os.Stat(exePath); err == nil {
				name := flatpakAppNames[app]
				browsers[name] = exePath
			}
		}
	}
}
