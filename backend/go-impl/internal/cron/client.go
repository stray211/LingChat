package cron

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron"
)

// Client 包裹gocron.Scheduler的客户端
type Client struct {
	scheduler    *gocron.Scheduler
	tempVoiceDir string
}

// NewClient 创建新的Cron客户端
func NewClient(tempVoiceDir string) *Client {
	// 使用服务器本地时区创建scheduler
	scheduler := gocron.NewScheduler(time.Local)

	return &Client{
		scheduler:    scheduler,
		tempVoiceDir: tempVoiceDir,
	}
}

// StartAsync 启动定时任务调度器并注册所有任务
func (c *Client) StartAsync() error {
	var (
		errFmt = "cron.StartAsync error:%s"
		tasks  []func() (*gocron.Job, error)
	)

	// 添加清理任务到任务列表
	tasks = append(tasks, c.registerCleanupTask)

	// 执行所有任务注册
	for _, task := range tasks {
		if _, err := task(); err != nil {
			return fmt.Errorf(errFmt, err)
		}
	}

	log.Println("Starting cron scheduler...")
	c.scheduler.StartAsync()
	return nil
}

// Stop 停止定时任务调度器
func (c *Client) Stop() {
	log.Println("Stopping cron scheduler...")
	c.scheduler.Stop()
}

// registerCleanupTask 注册清理任务，每小时40分执行
func (c *Client) registerCleanupTask() (*gocron.Job, error) {
	job, err := c.scheduler.Cron("40 * * * *").Do(c.cleanupOppositeDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to register cleanup task: %w", err)
	}

	log.Println("Registered cleanup task to run at 40 minutes of every hour")
	return job, nil
}

// cleanupOppositeDirectory 清理与当前小时奇偶性相反的目录
func (c *Client) cleanupOppositeDirectory() {
	currentHour := time.Now().Hour()
	var targetDir string

	// 判断当前小时奇偶性，删除相反的目录
	if currentHour%2 == 1 {
		// 当前小时是奇数，删除even目录
		targetDir = "even"
	} else {
		// 当前小时是偶数，删除odd目录
		targetDir = "odd"
	}

	targetPath := filepath.Join(c.tempVoiceDir, targetDir)

	log.Printf("Current hour: %d, cleaning up directory: %s", currentHour, targetDir)

	// 检查目录是否存在
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		log.Printf("Directory %s does not exist, skipping cleanup", targetPath)
		return
	}

	// 删除整个目录及其内容
	if err := os.RemoveAll(targetPath); err != nil {
		log.Printf("Failed to remove directory %s: %v", targetPath, err)
		return
	}

	log.Printf("Successfully cleaned up directory: %s", targetPath)
}
