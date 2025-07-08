package service

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"LingChat/internal/config"
)

// BackgroundService 背景服务接口
type BackgroundService interface {
	// ListBackgrounds 获取背景图片列表
	ListBackgrounds(ctx context.Context) ([]*BackgroundItem, error)
	// GetBackgroundFile 获取特定背景文件的完整路径
	GetBackgroundFile(ctx context.Context, backgroundFile string) (string, error)
}

// BackgroundItem 背景图片业务对象
type BackgroundItem struct {
	ImagePath    string  `json:"image_path"`
	Title        string  `json:"title"`
	ModifiedTime float64 `json:"modified_time"`
}

// backgroundService 背景服务实现
type backgroundService struct {
	config *config.Config
}

// NewBackgroundService 创建背景服务实例
func NewBackgroundService(config *config.Config) BackgroundService {
	return &backgroundService{
		config: config,
	}
}

// ListBackgrounds 获取背景图片列表
func (s *backgroundService) ListBackgrounds(ctx context.Context) ([]*BackgroundItem, error) {
	// 确定基础路径
	var basePath string
	if s.config != nil && s.config.Backend.ResourcePath != "" {
		basePath = s.config.Backend.ResourcePath
	} else {
		// 如果config不提供，使用当前工作目录
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		basePath = pwd
	}

	// 获取背景目录路径 - 使用config中的ResourcePath + game_data/backgrounds
	backgroundsDir := filepath.Join(basePath, "game_data", "backgrounds")

	// 检查目录是否存在
	if _, err := os.Stat(backgroundsDir); os.IsNotExist(err) {
		return []*BackgroundItem{}, nil // 返回空列表而不是错误
	}

	// 定义支持的图片格式
	supportedExtensions := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
	}

	// 读取目录中的文件
	files, err := os.ReadDir(backgroundsDir)
	if err != nil {
		return nil, err
	}

	var backgroundItems []*BackgroundItem
	for _, file := range files {
		// 跳过目录
		if file.IsDir() {
			continue
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !supportedExtensions[ext] {
			continue
		}

		// 获取文件信息
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		// 获取文件名（不含扩展名）
		title := strings.TrimSuffix(file.Name(), ext)

		// 创建背景项
		backgroundItem := &BackgroundItem{
			ImagePath:    file.Name(),
			Title:        title,
			ModifiedTime: float64(fileInfo.ModTime().Unix()),
		}

		backgroundItems = append(backgroundItems, backgroundItem)
	}

	// 按修改时间降序排序
	sort.Slice(backgroundItems, func(i, j int) bool {
		return backgroundItems[i].ModifiedTime > backgroundItems[j].ModifiedTime
	})

	return backgroundItems, nil
}

// GetBackgroundFile 获取特定背景文件的完整路径
func (s *backgroundService) GetBackgroundFile(ctx context.Context, backgroundFile string) (string, error) {
	// 确定基础路径
	var basePath string
	if s.config != nil && s.config.Backend.ResourcePath != "" {
		basePath = s.config.Backend.ResourcePath
	} else {
		// 如果config不提供，使用当前工作目录
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		basePath = pwd
	}

	// 构建完整的文件路径
	filePath := filepath.Join(basePath, "game_data", "backgrounds", backgroundFile)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}

	return filePath, nil
}
