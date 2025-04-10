package handle

import (
	"fmt"
	"log"
	"sync"
)

type EdgeBrowserInstance struct {
	browser Browser
	lock    sync.Mutex
}

var (
	edgeBrowserInstance *EdgeBrowserInstance
	edgeBrowseronce     sync.Once
)

func Edge() *EdgeBrowserInstance {
	edgeBrowseronce.Do(func() {
		edgeBrowserInstance = &EdgeBrowserInstance{}
	})
	return edgeBrowserInstance
}

func (m *EdgeBrowserInstance) Listen(debugPort int) (Browser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.browser != nil {
		if !m.browser.IsAlive() {
			m.browser.Close()
			m.browser = nil
		}
		return m.browser, nil
	}

	browsers := find_installed_browsers()
	if path, ok := browsers["edge"]; ok {
		log.Printf("Found Edge browser at: %s\n", path)
		browser, err := newEdgeBrowser(path, debugPort)
		if err != nil {
			return nil, fmt.Errorf("failed to start Edge browser: %w", err)
		}
		m.browser = browser
		return m.browser, nil
	}

	return nil, fmt.Errorf("edge browser not found")
}
