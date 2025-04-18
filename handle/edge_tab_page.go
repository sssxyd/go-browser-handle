package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/playwright-community/playwright-go"
)

type EdgeTabPage struct {
	id      string          // 标签页ID
	url     string          // 标签页初始URL
	browser *EdgeBrowser    // 浏览器实例
	page    playwright.Page // 标签页实例
}

func newEdgeTabPage(id string, url string, browser *EdgeBrowser, page playwright.Page) *EdgeTabPage {

	tabPage := &EdgeTabPage{
		id:      id,
		url:     url,
		browser: browser,
		page:    page,
	}

	return tabPage
}

func (t *EdgeTabPage) ID() string {
	return t.id
}

func (t *EdgeTabPage) Title() string {
	title, err := t.page.Title()
	if err != nil {
		log.Printf("Failed to get page title: %v", err)
		return ""
	}
	return title
}

func (t *EdgeTabPage) URL() string {
	t.page.BringToFront()
	return t.page.URL()
}

func (t *EdgeTabPage) Domain() string {
	url := t.URL()
	domain, err := extract_domain_from_url(url)
	if err != nil {
		log.Printf("Failed to extract domain from URL: %s", url)
		return ""
	}
	return domain
}

func (t *EdgeTabPage) IsClosed() bool {
	return t.page.IsClosed()
}

func (t *EdgeTabPage) BringToFront() {
	t.page.BringToFront()
}

func (t *EdgeTabPage) Page() playwright.Page {
	return t.page
}

func (t *EdgeTabPage) OpenInNewTab(id string, action func() error, timeout float64) TabPage {
	t.browser.locker.Lock()
	defer t.browser.locker.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	newPageChan := make(chan playwright.Page, 1)
	done := make(chan struct{}) // 用于等待 goroutine 退出

	go func() {
		defer close(done) // 确保 goroutine 退出时发送信号

		select {
		case <-ctx.Done():
			return
		default:
		}

		newPage, err := t.browser.context.WaitForEvent("page", playwright.BrowserContextWaitForEventOptions{
			Predicate: func(event any) bool { return true },
			Timeout:   playwright.Float(timeout),
		})
		if err != nil {
			log.Printf("等待新页面失败: %v", err)
			return
		}

		newPageObj := newPage.(playwright.Page)
		if err := newPageObj.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		}); err != nil {
			newPageObj.Close()
			return
		}

		listen_page_console_log(newPageObj)
		block_debug_port_detector(newPageObj, t.browser.port)

		select {
		case newPageChan <- newPageObj:
		case <-ctx.Done():
			newPageObj.Close()
		}
	}()

	if err := action(); err != nil {
		cancel() // 取消上下文
		<-done   // 等待 goroutine 退出
		return nil
	}

	select {
	case newPage := <-newPageChan:
		tabPage := t.browser.addTabPage(id, newPage.URL(), newPage)
		log.Printf("成功捕获新标签页, ID: %s", id)
		return tabPage
	case <-ctx.Done():
		<-done // 等待 goroutine 清理完毕
		log.Printf("等待新标签页超时: %v", timeout)
		return nil
	}
}

func (t *EdgeTabPage) WaitSelector(selector string, timeout float64) playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return nil
	}
	// await expect(lolocator).toHaveCount(1)
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return nil
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return nil
	}
	if count > 1 {
		log.Printf("选择器匹配到多个元素: %s, 取第一条返回", selector)
		locator = locator.First()
	}
	err = locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(timeout),
	})
	if err != nil {
		return nil
	}
	return locator
}

func (t *EdgeTabPage) QuerySelector(selector string) playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return nil
	}
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return nil
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return nil
	}
	if count > 1 {
		log.Printf("选择器匹配到多个元素: %s, 取第一条返回", selector)
		locator = locator.First()
	}
	return locator
}

func (t *EdgeTabPage) QuerySelectorAll(selector string) []playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return []playwright.Locator{}
	}
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return []playwright.Locator{}
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return []playwright.Locator{}
	}
	if count == 1 {
		return []playwright.Locator{locator}
	}
	items, err := locator.All()
	if err != nil {
		log.Printf("无法获取所有选择器: %v", err)
		return []playwright.Locator{}
	}
	return items
}

func (t *EdgeTabPage) ClearLocalData() error {
	log.Printf("正在清除站点%s的所有本地存储...", t.Domain())
	if _, err := t.page.Evaluate("localStorage.clear()"); err != nil {
		return fmt.Errorf("清空 localStorage 失败: %w", err)
	}
	if _, err := t.page.Evaluate("sessionStorage.clear()"); err != nil {
		return fmt.Errorf("清空 sessionStorage 失败: %w", err)
	}
	if err := t.browser.context.ClearCookies(); err != nil {
		return fmt.Errorf("清除 Cookies 失败: %w", err)
	}
	if _, err := t.page.Evaluate(`
        async () => {
            const databases = await window.indexedDB.databases();
            for (const db of databases) {
                if (db.name) {
                    window.indexedDB.deleteDatabase(db.name);
                }
            }
        }
    `); err != nil {
		return fmt.Errorf("清除 IndexedDB 失败: %w", err)
	}
	return nil
}

func (t *EdgeTabPage) Goto(url string) error {
	// 导航
	_, err := t.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return fmt.Errorf("无法访问网站: %v", err)
	}

	block_debug_port_detector(t.page, t.browser.port)

	// 等待页面完全加载
	err = t.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
	if err != nil {
		return fmt.Errorf("等待页面加载失败: %v", err)
	}
	log.Printf("已成功访问网站: %s", url)

	return nil
}

func (t *EdgeTabPage) Evaluate(expression string, arg ...any) (any, error) {
	return t.page.Evaluate(expression, arg...)
}

func (t *EdgeTabPage) Close() {
	t.browser.removeTabPage(t.id)
}

func (t *EdgeTabPage) Reload() error {
	return t.Goto(t.URL())
}

func (t *EdgeTabPage) GetCookies() string {
	cookies, err := t.page.Context().Cookies()
	if err != nil {
		return ""
	}
	bytes, err := json.Marshal(cookies)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (t *EdgeTabPage) ApplyCookies(cookies string) error {
	var cookieList []playwright.OptionalCookie
	if err := json.Unmarshal([]byte(cookies), &cookieList); err != nil {
		return fmt.Errorf("无法解析 Cookies: %w", err)
	}
	// 先删除原有的 Cookies
	if err := t.page.Context().ClearCookies(); err != nil {
		log.Printf("无法清除 Cookies: %v", err)
		return fmt.Errorf("无法清除 Cookies: %w", err)
	}
	if err := t.page.Context().AddCookies(cookieList); err != nil {
		return fmt.Errorf("无法设置 Cookies: %w", err)
	}
	return nil
}

func (t *EdgeTabPage) SleepRandom(min, max int) {
	if min < 0 || max < 0 || min > max {
		log.Printf("无效的随机时间范围: %d - %d", min, max)
		return
	}
	sleepTime := min + rand.Intn(max-min+1)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
