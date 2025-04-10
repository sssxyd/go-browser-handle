package test

import (
	"testing"

	"github.com/sssxyd/go-browser-handle/src/handle" // 导入 src/handle 包
)

func TestRemoveByValue(t *testing.T) {
	edge, err := handle.StartEdgeBrowser("edgePath", 3000, 4000) // 使用 handle 包中的函数
	if err != nil {
		t.Fatalf("Failed to start Edge browser: %v", err)
	}
	_ = edge // 测试逻辑
}
