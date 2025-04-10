package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
)

var (
	edgeBrowser Browser
	edgeLock    sync.Mutex
)

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

func StartEdgeBrowser(edgePath string, debugPort int) (Browser, error) {
	browser, err := newEdgeBrowser(edgePath, debugPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start Edge browser: %w", err)
	}
	return browser, nil
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

func GetEdgeBrowser(debugPort int) (Browser, error) {
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
		browser, err := StartEdgeBrowser(path, debugPort)
		if err != nil {
			return nil, fmt.Errorf("failed to start Edge browser: %w", err)
		}
		edgeBrowser = browser
		return edgeBrowser, nil
	}

	return nil, fmt.Errorf("edge browser not found")
}
