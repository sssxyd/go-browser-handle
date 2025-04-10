package handle

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func connect_to_exist_edge(pw *playwright.Playwright, port string) (playwright.Browser, bool) {
	// 尝试连接到默认调试端口
	browser, err := pw.Chromium.ConnectOverCDP("http://127.0.0.1:" + port)
	if err == nil {
		return browser, true
	}
	return nil, false
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

func extract_domain_from_url(urlStr string) (string, error) {
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
