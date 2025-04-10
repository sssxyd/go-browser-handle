package handle

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Windows实现
func _findWindowsBrowsers(browsers map[string]string) {
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
func _findMacBrowsers(browsers map[string]string) {
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
func _findLinuxBrowsers(browsers map[string]string) {
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

func find_installed_browsers() map[string]string {
	browsers := make(map[string]string)

	switch runtime.GOOS {
	case "windows":
		_findWindowsBrowsers(browsers)
	case "darwin":
		_findMacBrowsers(browsers)
	case "linux":
		_findLinuxBrowsers(browsers)
	}
	return browsers
}

func kill_edge_processes() error {
	// 检查进程是否存在
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq msedge.exe")
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "msedge.exe") {
		fmt.Println("未找到 msedge.exe 进程")
		return nil
	}

	// 终止进程
	cmd = exec.Command("taskkill", "/F", "/IM", "msedge.exe")
	output, err = cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			log.Printf("命令退出码: %d, 输出: %s", exitCode, string(output))
			switch exitCode {
			case 1:
				return fmt.Errorf("权限不足，请以管理员身份运行")
			case 128:
				return fmt.Errorf("未找到 msedge.exe 进程")
			default:
				return fmt.Errorf("未知错误，退出码: %d, 输出: %s", exitCode, string(output))
			}
		}
		return fmt.Errorf("终止 msedge 失败: %v, 输出: %s", err, string(output))
	}
	fmt.Println("成功终止 msedge 进程")
	return nil
}
