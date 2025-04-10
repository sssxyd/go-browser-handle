package test

import (
	"testing"

	"github.com/sssxyd/go-browser-handle/handle"
)

func TestRemoveByValue(t *testing.T) {
	edge, err := handle.StartEdgeBrowser("edgePath", 3000, 4000)
	if err != nil {
		t.Fatalf("Failed to start Edge browser: %v", err)
	}
	_ = edge // 测试逻辑
}
