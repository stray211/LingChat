package service

import (
	"testing"
	"time"
)

func TestVoiceStatusTracker(t *testing.T) {
	// 创建状态跟踪器
	tracker, err := NewVoiceStatusTracker()
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	filename := "test_voice_file.wav"

	// 测试设置pending状态
	err = tracker.SetStatus(filename, StatusPending)
	if err != nil {
		t.Errorf("Failed to set pending status: %v", err)
	}

	// 测试获取状态
	status := tracker.GetStatus(filename)
	if status.Status != StatusPending {
		t.Errorf("Expected status %s, got %s", StatusPending, status.Status)
	}

	// 测试更新状态为ready
	err = tracker.UpdateStatus(filename, StatusReady)
	if err != nil {
		t.Errorf("Failed to update status to ready: %v", err)
	}

	// 再次获取状态验证
	status = tracker.GetStatus(filename)
	if status.Status != StatusReady {
		t.Errorf("Expected status %s, got %s", StatusReady, status.Status)
	}

	// 测试不存在的文件
	nonExistentStatus := tracker.GetStatus("non_existent_file.wav")
	if nonExistentStatus.Status != StatusInvalid {
		t.Errorf("Expected status %s for non-existent file, got %s", StatusInvalid, nonExistentStatus.Status)
	}

	if !nonExistentStatus.CreateAt.IsZero() {
		t.Errorf("Expected zero time for non-existent file, got %v", nonExistentStatus.CreateAt)
	}
}

func TestVoiceStatusTracker_TTL(t *testing.T) {
	// Note: 这个测试需要等待实际时间，在单元测试中不太适合运行
	// 仅作为功能演示，实际项目可能需要使用时间模拟
	t.Skip("Skipping TTL test as it requires actual time to pass")

	tracker, err := NewVoiceStatusTracker()
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	filename := "ttl_test_file.wav"

	// 设置状态
	err = tracker.SetStatus(filename, StatusPending)
	if err != nil {
		t.Errorf("Failed to set status: %v", err)
	}

	// 等待TTL过期 (20分钟太长，这里只是示例)
	time.Sleep(1 * time.Second)

	// 检查状态是否仍然存在
	status := tracker.GetStatus(filename)
	t.Logf("Status after wait: %s", status.Status)
}
