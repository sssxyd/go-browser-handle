package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/playwright-community/playwright-go"
	"golang.org/x/sys/windows/registry"
)

func removeByValue(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	//
	return slice
}

func extractDomainFromUrl(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "://") { // 补全缺失的协议头
		urlStr = "http://" + urlStr
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	hostname := u.Hostname()
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid hostname: %s", hostname)
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1], nil
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func getValidPort(start, end int) (int, error) {
	count := 0
	for i := start; i <= end; i++ {
		// 随机选择一个端口
		port := i
		// 检查端口是否可用
		if isPortAvailable(port) {
			return port, nil
		}
		count++
		fmt.Printf("端口 %d 被占用，尝试次数 %d/%d\n", port, i+1, count)
	}
	return 0, fmt.Errorf("在 %d 次尝试后未找到可用端口", count)
}

func connectToExistingEdge(pw *playwright.Playwright, port string) (playwright.Browser, bool) {
	// 尝试连接到默认调试端口
	browser, err := pw.Chromium.ConnectOverCDP("http://127.0.0.1:" + port)
	if err == nil {
		return browser, true
	}
	return nil, false
}

func killEdgeProcesses() error {
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

// startNewEdge 启动新的 Edge 实例
func startNewEdge(edgePath, port string) (*exec.Cmd, error) {
	cmd := exec.Command(edgePath,
		"--new-window",
		"about:blank",
		"--remote-debugging-port="+port,
		"--remote-allow-origins=http://127.0.0.1:"+port)
	log.Printf("启动 Edge 浏览器: %s", cmd.String())
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("无法启动 Edge 浏览器: %v", err)
	}
	log.Printf("Edge 浏览器已启动，调试端口: %s", port)
	return cmd, nil
}

func block_debug_port_detector(p playwright.Page, port int) error {
	script := fmt.Sprintf(`() => {
		const debugPort = "%d";
		const originalGetEntries = performance.getEntries;
		const originalGetEntriesByType = performance.getEntriesByType;
		const OriginalWebSocket = window.WebSocket;

		// 过滤 Performance 条目
        performance.getEntries = function() {
            return originalGetEntries.call(this).filter(entry => 
                !entry.name.includes("127.0.0.1:" + debugPort) && 
                !entry.name.includes("localhost:" + debugPort)
            );
        };

		performance.getEntriesByType = function(type){
			return originalGetEntriesByType.call(this, type).filter(entry =>
                !entry.name.includes("127.0.0.1:" + debugPort) && 
                !entry.name.includes("localhost:" + debugPort)
			);
		};

		// 拦截 WebSocket 连接
		window.WebSocket = function(urlArg, protocols) {
		// 统一处理 URL 格式
		const url = urlArg instanceof URL ? urlArg.href : urlArg;
		
		// 检测目标地址
		if (typeof url === 'string' && 
			(url.includes("127.0.0.1:" + debugPort) || 
			url.includes("localhost:" + debugPort))) {
			
			// 创建虚假的 WebSocket 对象
			const fakeWs = new OriginalWebSocket('ws://invalid-host-' + Date.now());
			
			// 立即关闭连接并修改状态
			Object.defineProperty(fakeWs, 'readyState', {
			value: OriginalWebSocket.CLOSED,
			writable: false
			});
			
			// 强制触发错误事件
			const errorEvent = new Event('error');
			errorEvent.initEvent('error', false, false);
			
			// 设置异步触发确保执行顺序
			setTimeout(() => {
			if (typeof fakeWs.onerror === 'function') {
				fakeWs.onerror(errorEvent);
			}
			fakeWs.dispatchEvent(errorEvent);
			
			fakeWs.close();
			}, 0);
			
			return fakeWs;
		}
		
		// 正常连接处理
		return protocols ? 
			new OriginalWebSocket(urlArg, protocols) : 
			new OriginalWebSocket(urlArg);
		};

		// 保持原型链完整
		window.WebSocket.prototype = OriginalWebSocket.prototype;		
	}`, port)
	_, err := p.Evaluate(script)
	if err != nil {
		log.Printf("注入拦截脚本失败: %v", err)
		return err
	}

	log.Printf(">>>已拦截Performance API和 WebSocket 请求")

	return nil
}

func listen_page_console_log(page playwright.Page) {
	page.On("console", func(message playwright.ConsoleMessage) {
		log.Printf(">>> %s", message.Text())
	})
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

func ParseJson[T any](jsonstr string) (T, error) {
	var obj T
	err := json.Unmarshal([]byte(jsonstr), &obj)
	if err != nil {
		return obj, fmt.Errorf("json unmarshal error: %w", err)
	}
	return obj, nil
}

func Stringify[T any](obj T) (string, error) {
	// JSON序列化
	jsonstr, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %w", err)
	}
	return string(jsonstr), nil
}

func FindInstalledBrowsers() map[string]string {
	browsers := make(map[string]string)

	switch runtime.GOOS {
	case "windows":
		findWindowsBrowsers(browsers)
	case "darwin":
		findMacBrowsers(browsers)
	case "linux":
		findLinuxBrowsers(browsers)
	}
	return browsers
}
