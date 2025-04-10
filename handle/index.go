package handle

import (
	"fmt"
	"log"
	"sync"
)

var (
	edgeBrowser Browser
	edgeLock    sync.Mutex
)

func StartEdgeBrowser(debugPort int) (Browser, error) {
	edgeLock.Lock()
	defer edgeLock.Unlock()

	if edgeBrowser != nil {
		if !edgeBrowser.IsAlive() {
			edgeBrowser.Close()
			edgeBrowser = nil
		}
		return edgeBrowser, nil
	}

	browsers := FindInstalledBrowsers()
	if path, ok := browsers["edge"]; ok {
		log.Printf("Found Edge browser at: %s\n", path)
		browser, err := newEdgeBrowser(path, debugPort)
		if err != nil {
			return nil, fmt.Errorf("failed to start Edge browser: %w", err)
		}
		edgeBrowser = browser
		return edgeBrowser, nil
	}

	return nil, fmt.Errorf("edge browser not found")
}
