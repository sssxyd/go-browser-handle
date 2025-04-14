package handle

import (
	"github.com/playwright-community/playwright-go"
)

type TabPage interface {
	ID() string
	Title() string
	URL() string
	Domain() string
	Close()
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
	Reload() error
	GetCookies() string
	ApplyCookies(cookies string) error
	SleepRandom(min, max int)
}

type Browser interface {
	Name() string
	Port() int
	TabPages() []TabPage
	NewTabPage(id string, url string) TabPage
	DefaultPage() TabPage
	FindTabPage(id string) TabPage
	SwitchToTabPage(id string) error
	CloseTabPage(id string) error
	IsAlive() bool
	Close() error
}
