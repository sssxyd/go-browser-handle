package handle_test

import (
	"log"
	"testing"
	"time"

	"github.com/sssxyd/go-browser-handle/handle"
)

var (
	port    int = 9812
	browser handle.Browser
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // 时间戳 + 文件名行号

	edge, err := handle.Edge().Listen(port)
	if err != nil {
		panic(err)
	}
	browser = edge
}

func dispose() {
	if browser != nil {
		browser.Close()
	}
}

func TestVisitYFW(t *testing.T) {
	t.Cleanup(dispose)

	page := browser.DefaultPage()
	if page == nil {
		t.Fatalf("未找到默认标签页")
	}

	page.BringToFront()

	time.Sleep(5 * time.Second)
	if err := page.Goto("https://www.yaofangwang.com/"); err != nil {
		t.Fatalf("访问页面失败: %v", err)
	}
	time.Sleep(5 * time.Second)
	page.Close()
	time.Sleep(2 * time.Second)
	browser.Close()
}
