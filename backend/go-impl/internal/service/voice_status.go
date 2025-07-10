package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/allegro/bigcache/v3"
)

// VoiceFileStatus 语音文件状态
type VoiceFileStatus struct {
	CreateAt time.Time `json:"create_at"`
	Status   string    `json:"status"`
}

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
	StatusInvalid = "Invalid"
)

// VoiceStatusTracker 语音文件状态跟踪器
type VoiceStatusTracker struct {
	cache *bigcache.BigCache
}

// NewVoiceStatusTracker 创建新的状态跟踪器
func NewVoiceStatusTracker() (*VoiceStatusTracker, error) {
	config := bigcache.DefaultConfig(20 * time.Minute) // TTL 20分钟
	config.CleanWindow = 5 * time.Minute               // 清理窗口5分钟
	config.MaxEntriesInWindow = 1000 * 10 * 60         // 10分钟内最大条目数
	config.MaxEntrySize = 500                          // 最大条目大小
	config.Verbose = false                             // 关闭详细日志

	cache, err := bigcache.New(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &VoiceStatusTracker{
		cache: cache,
	}, nil
}

// SetStatus 设置文件状态
func (vst *VoiceStatusTracker) SetStatus(filename string, status string) error {
	voiceStatus := VoiceFileStatus{
		CreateAt: time.Now(),
		Status:   status,
	}

	data, err := json.Marshal(voiceStatus)
	if err != nil {
		return err
	}

	return vst.cache.Set(filename, data)
}

// GetStatus 获取文件状态
func (vst *VoiceStatusTracker) GetStatus(filename string) VoiceFileStatus {
	data, err := vst.cache.Get(filename)
	if err != nil {
		// 键不存在或其他错误，返回Invalid状态
		return VoiceFileStatus{
			CreateAt: time.Time{}, // 零值
			Status:   StatusInvalid,
		}
	}

	var status VoiceFileStatus
	if err := json.Unmarshal(data, &status); err != nil {
		// JSON解析错误，返回Invalid状态
		return VoiceFileStatus{
			CreateAt: time.Time{}, // 零值
			Status:   StatusInvalid,
		}
	}

	return status
}

// UpdateStatus 更新文件状态（保持原始创建时间）
func (vst *VoiceStatusTracker) UpdateStatus(filename string, newStatus string) error {
	// 先获取现有状态
	currentStatus := vst.GetStatus(filename)
	if currentStatus.Status == StatusInvalid {
		// 如果不存在，创建新的
		return vst.SetStatus(filename, newStatus)
	}

	// 更新状态但保持原始创建时间
	updatedStatus := VoiceFileStatus{
		CreateAt: currentStatus.CreateAt,
		Status:   newStatus,
	}

	data, err := json.Marshal(updatedStatus)
	if err != nil {
		return err
	}

	return vst.cache.Set(filename, data)
}

// Close 关闭缓存
func (vst *VoiceStatusTracker) Close() error {
	return vst.cache.Close()
}
