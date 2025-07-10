package cron

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupOppositeDirectory(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	// 创建odd和even子目录以及测试文件
	oddDir := filepath.Join(tempDir, "odd")
	evenDir := filepath.Join(tempDir, "even")

	if err := os.MkdirAll(oddDir, 0755); err != nil {
		t.Fatalf("Failed to create odd directory: %v", err)
	}
	if err := os.MkdirAll(evenDir, 0755); err != nil {
		t.Fatalf("Failed to create even directory: %v", err)
	}

	// 在each目录中创建测试文件
	oddFile := filepath.Join(oddDir, "test_odd.wav")
	evenFile := filepath.Join(evenDir, "test_even.wav")

	if err := os.WriteFile(oddFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file in odd directory: %v", err)
	}
	if err := os.WriteFile(evenFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file in even directory: %v", err)
	}

	// 创建cron client
	client := NewClient(tempDir)

	// 获取当前小时并验证清理逻辑
	currentHour := time.Now().Hour()
	var shouldRemainDir, shouldBeRemovedDir string

	if currentHour%2 == 1 {
		// 当前小时是奇数，应该删除even目录，保留odd目录
		shouldRemainDir = oddDir
		shouldBeRemovedDir = evenDir
	} else {
		// 当前小时是偶数，应该删除odd目录，保留even目录
		shouldRemainDir = evenDir
		shouldBeRemovedDir = oddDir
	}

	// 执行清理
	client.cleanupOppositeDirectory()

	// 验证结果
	if _, err := os.Stat(shouldRemainDir); os.IsNotExist(err) {
		t.Errorf("Directory %s should remain but was deleted", shouldRemainDir)
	}

	if _, err := os.Stat(shouldBeRemovedDir); !os.IsNotExist(err) {
		t.Errorf("Directory %s should be deleted but still exists", shouldBeRemovedDir)
	}

	t.Logf("Test passed: current hour %d, removed directory %s, kept directory %s",
		currentHour, filepath.Base(shouldBeRemovedDir), filepath.Base(shouldRemainDir))
}

func TestStartAsync(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	// 创建cron client
	client := NewClient(tempDir)

	// 测试StartAsync
	if err := client.StartAsync(); err != nil {
		t.Fatalf("StartAsync failed: %v", err)
	}

	// 停止scheduler
	client.Stop()

	t.Log("StartAsync test passed")
}
