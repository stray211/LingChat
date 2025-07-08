package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"LingChat/internal/config"
	"LingChat/internal/data"
	"LingChat/internal/data/ent/ent"
)

// CharacterService 角色服务接口
type CharacterService interface {
	// GetAllCharacters 获取所有角色列表
	GetAllCharacters(ctx context.Context) ([]*CharacterItem, error)
	// GetCharacterByID 根据ID获取角色
	GetCharacterByID(ctx context.Context, id int64) (*CharacterItem, error)
	// GetCharacterByCharacterID 根据character_id获取角色
	GetCharacterByCharacterID(ctx context.Context, characterID string) (*CharacterItem, error)
	// GetCharacterAvatarPath 获取角色头像文件的完整路径
	GetCharacterAvatarPath(ctx context.Context, characterID string, avatarFile string) (string, error)
	// GetCharacterInfo 根据character_id获取角色详细信息
	GetCharacterInfo(ctx context.Context, characterID string) (*CharacterInfoItem, error)
}

// CharacterItem 角色业务对象
type CharacterItem struct {
	CharacterID  string `json:"character_id"`
	Title        string `json:"title"`
	Info         string `json:"info"`
	AvatarPath   string `json:"avatar_path"`
	ResourcePath string `json:"resource_path"`
}

// CharacterInfoItem 角色详细信息业务对象
type CharacterInfoItem struct {
	AIName          string `json:"ai_name"`
	AISubtitle      string `json:"ai_subtitle"`
	UserName        string `json:"user_name"`
	UserSubtitle    string `json:"user_subtitle"`
	CharacterID     string `json:"character_id"`
	ThinkingMessage string `json:"thinking_message"`
	Scale           string `json:"scale"`
	Offset          string `json:"offset"`
	BubbleTop       string `json:"bubble_top"`
	BubbleLeft      string `json:"bubble_left"`
}

// characterService 角色服务实现
type characterService struct {
	characterRepo data.CharacterRepo
	config        *config.Config
}

// NewCharacterService 创建角色服务实例
func NewCharacterService(characterRepo data.CharacterRepo, config *config.Config) CharacterService {
	return &characterService{
		characterRepo: characterRepo,
		config:        config,
	}
}

// GetAllCharacters 获取所有角色列表
func (s *characterService) GetAllCharacters(ctx context.Context) ([]*CharacterItem, error) {
	// 从数据库获取所有角色
	characters, _, err := s.characterRepo.List(ctx, 0, 1000) // 获取前1000个角色
	if err != nil {
		return nil, fmt.Errorf("failed to get characters from database: %w", err)
	}

	// 转换为业务对象
	var items []*CharacterItem
	for _, char := range characters {
		item := s.convertToCharacterItem(char)
		items = append(items, item)
	}

	return items, nil
}

// GetCharacterByID 根据ID获取角色
func (s *characterService) GetCharacterByID(ctx context.Context, id int64) (*CharacterItem, error) {
	char, err := s.characterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get character by id: %w", err)
	}

	return s.convertToCharacterItem(char), nil
}

// GetCharacterByCharacterID 根据character_id获取角色
func (s *characterService) GetCharacterByCharacterID(ctx context.Context, characterID string) (*CharacterItem, error) {
	char, err := s.characterRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character by character_id: %w", err)
	}

	return s.convertToCharacterItem(char), nil
}

// GetCharacterAvatarPath 获取角色头像文件的完整路径
func (s *characterService) GetCharacterAvatarPath(ctx context.Context, characterID string, avatarFile string) (string, error) {
	// 查找角色
	character, err := s.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		return "", fmt.Errorf("failed to get character by character_id: %w", err)
	}

	// 确定基础路径
	var basePath string
	if s.config != nil && s.config.Backend.ResourcePath != "" {
		basePath = s.config.Backend.ResourcePath
	} else {
		// 如果config不提供，使用当前工作目录
		pwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = pwd
	}

	// 构建完整的文件路径: basePath + character.ResourcePath + avatarFile
	fullPath := filepath.Join(basePath, character.ResourcePath, "avatar", avatarFile)

	return fullPath, nil
}

// GetCharacterInfo 根据character_id获取角色详细信息
func (s *characterService) GetCharacterInfo(ctx context.Context, characterID string) (*CharacterInfoItem, error) {
	char, err := s.characterRepo.GetByCharacterID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character by character_id: %w", err)
	}

	// 解析display_params
	displayParams := make(map[string]interface{})
	if char.DisplayParams != nil {
		for k, v := range char.DisplayParams {
			displayParams[k] = v
		}
	}

	// 辅助函数：安全地从map中获取字符串值，支持多种类型转换
	getString := func(key string) string {
		if v, ok := displayParams[key]; ok {
			switch val := v.(type) {
			case string:
				return val
			case int:
				return fmt.Sprintf("%d", val)
			case int64:
				return fmt.Sprintf("%d", val)
			case float64:
				return fmt.Sprintf("%g", val)
			case float32:
				return fmt.Sprintf("%g", val)
			default:
				return fmt.Sprintf("%v", val)
			}
		}
		return ""
	}

	return &CharacterInfoItem{
		AIName:          getString("ai_name"),
		AISubtitle:      getString("ai_subtitle"),
		UserName:        getString("user_name"),
		UserSubtitle:    getString("user_subtitle"),
		CharacterID:     char.CharacterID,
		ThinkingMessage: getString("thinking_message"),
		Scale:           getString("scale"),
		Offset:          getString("offset"),
		BubbleTop:       getString("bubble_top"),
		BubbleLeft:      getString("bubble_left"),
	}, nil
}

// convertToCharacterItem 将数据库实体转换为业务对象
func (s *characterService) convertToCharacterItem(char *ent.Character) *CharacterItem {
	// 构造头像路径
	avatarPath := s.buildAvatarPath(char.ResourcePath)

	// 如果数据库中的info为空，提供默认值
	info := char.Info
	if info == "" {
		info = "这是一个人工智能对话助手"
	}

	return &CharacterItem{
		CharacterID:  char.CharacterID,
		Title:        char.Title,
		Info:         info,
		AvatarPath:   avatarPath,
		ResourcePath: char.ResourcePath,
	}
}

// buildAvatarPath 构造头像路径
func (s *characterService) buildAvatarPath(resourcePath string) string {
	if resourcePath == "" {
		return ""
	}

	// 提取角色名
	characterName := filepath.Base(resourcePath)
	if characterName == "" || characterName == "." {
		return ""
	}

	// 构造相对路径: "角色名/avatar/正常.png"
	return fmt.Sprintf("%s/avatar/正常.png", characterName)
}
