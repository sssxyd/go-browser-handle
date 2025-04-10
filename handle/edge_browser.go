package handle

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"slices"

	"github.com/playwright-community/playwright-go"
)

type EdgeBrowser struct {
	pw       *playwright.Playwright
	port     int
	browser  playwright.Browser
	context  playwright.BrowserContext
	tabPages []*EdgeTabPage
	locker   sync.Mutex
}

// startNewEdge 启动新的 Edge 实例
func startNewEdge(edgePath, port string) (*exec.Cmd, error) {
	cmd := exec.Command(edgePath,
		"--new-window",
		"about:blank",
		"--remote-debugging-address=127.0.0.1",
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

func newEdgeBrowser(edgePath string, debugPort int) (*EdgeBrowser, error) {
	// 1. 自动安装 Playwright 驱动
	if err := playwright.Install(); err != nil {
		return nil, err
	}

	// 2. 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("无法启动 Playwright: %v", err)
	}

	// 3. 尝试连接到已运行的 Edge 实例
	browser, found := connect_to_exist_edge(pw, fmt.Sprintf("%d", debugPort))
	if !found {
		// 如果没有找到已运行的 Edge 实例，则关闭所有 Edge 进程并启动一个新实例
		err = kill_edge_processes()
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("无法关闭 Edge 进程: %v", err)
		}

		// 启动新的 Edge 实例
		_, err = startNewEdge(edgePath, fmt.Sprintf("%d", debugPort))
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("无法启动 Edge 浏览器: %v", err)
		}

		// 等待片刻，确保浏览器启动
		time.Sleep(2 * time.Second)
		log.Printf("Edge 浏览器已通过系统命令启动，调试端口: %d\n", debugPort)

		// 连接到新启动的 Edge 实例
		browser, found = connect_to_exist_edge(pw, fmt.Sprintf("%d", debugPort))
		if !found {
			pw.Stop()
			return nil, fmt.Errorf("无法连接到新启动的 Edge 实例")
		}
	} else {
		log.Printf("已找到正在运行的 Edge 实例，调试端口:%d\n", debugPort)
	}

	// 4. 获取浏览器上下文并关闭所有默认页面
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("未找到任何浏览器上下文")
	}
	browserContext := contexts[0]
	pages := browserContext.Pages()
	for _, p := range pages {
		err = p.Close()
		if err != nil {
			log.Printf("关闭页面失败: %v", err)
		}
	}

	// 5. 创建 EdgeBrowser 实例
	pe := &EdgeBrowser{
		port:     debugPort,
		pw:       pw,
		browser:  browser,
		context:  browserContext,
		tabPages: make([]*EdgeTabPage, 0),
		locker:   sync.Mutex{},
	}

	// 6. 创建默认标签页
	tabPage := pe.NewTabPage("default", "about:blank")
	if tabPage == nil {
		pe.Close()
		return nil, fmt.Errorf("无法创建默认标签页")
	}

	return pe, nil
}

func (b *EdgeBrowser) addTabPage(id string, url string, page playwright.Page) *EdgeTabPage {
	tabPage := newEdgeTabPage(id, url, b, page)
	b.tabPages = append(b.tabPages, tabPage)
	return tabPage
}

func (b *EdgeBrowser) removeTabPage(id string) {
	var tabPage *EdgeTabPage
	for i, page := range b.tabPages {
		if page.ID() == id {
			tabPage = b.tabPages[i]
			b.tabPages = slices.Delete(b.tabPages, i, i+1)
			break
		}
	}

	if tabPage != nil {
		if !tabPage.page.IsClosed() {
			tabPage.page.Close()
		}
	}
}

func (b *EdgeBrowser) Name() string {
	return "edge"
}

func (b *EdgeBrowser) Port() int {
	return b.port
}

func (b *EdgeBrowser) IsAlive() bool {
	return b.browser.IsConnected()
}

func (b *EdgeBrowser) NewTabPage(id string, url string) TabPage {
	b.locker.Lock()
	defer b.locker.Unlock()

	if url == "" {
		url = "about:blank"
	}
	// 创建一个新的空白页面
	page, err := b.context.NewPage()
	if err != nil {
		log.Printf("无法创建新页面: %v", err)
		return nil
	}

	// 监听控制台消息
	listen_page_console_log(page)

	tabPage := b.addTabPage(id, url, page)

	err = tabPage.Goto(url)
	if err != nil {
		b.removeTabPage(tabPage.id)
		log.Printf("无法打开页面: %v", err)
		return nil
	}

	return tabPage
}

func (b *EdgeBrowser) DefaultPage() TabPage {
	return b.FindTabPage("default")
}

func (b *EdgeBrowser) FindTabPage(id string) TabPage {
	for _, page := range b.tabPages {
		if page.ID() == id {
			return page
		}
	}
	return nil
}

func (b *EdgeBrowser) TabPages() []TabPage {
	var tabPages []TabPage = make([]TabPage, 0, len(b.tabPages))
	for _, page := range b.tabPages {
		tabPages = append(tabPages, page)
	}
	return tabPages
}

func (b *EdgeBrowser) SwitchToTabPage(id string) error {
	tabPage := b.FindTabPage(id)
	if tabPage == nil {
		return fmt.Errorf("未找到标签页: %s", id)
	}

	tabPage.BringToFront()
	return nil
}

func (b *EdgeBrowser) CloseTabPage(id string) error {
	b.locker.Lock()
	defer b.locker.Unlock()

	tabPage := b.FindTabPage(id)
	if tabPage == nil {
		return fmt.Errorf("未找到标签页: %s", id)
	}

	b.removeTabPage(id)
	return nil
}

func (b *EdgeBrowser) Close() error {
	b.locker.Lock()
	defer b.locker.Unlock()

	for _, page := range b.tabPages {
		if !page.page.IsClosed() {
			page.page.Close()
		}
	}

	err := b.browser.Close()
	if err != nil {
		log.Printf("关闭浏览器失败: %v", err)
		return err
	}

	b.pw.Stop()
	return nil
}
