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
}

// CharacterItem 角色业务对象
type CharacterItem struct {
	CharacterID  string `json:"character_id"`
	Title        string `json:"title"`
	Info         string `json:"info"`
	AvatarPath   string `json:"avatar_path"`
	ResourcePath string `json:"resource_path"`
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
